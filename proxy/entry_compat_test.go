package proxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/lqqyt2423/go-mitmproxy/cert"
	"github.com/mastercactapus/proxyprotocol"
)

func TestCustomEntryPreservesClosingResponse(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tls, pp bool
	}{
		{"http", false, false},
		{"https", true, false},
		{"proxy_protocol", false, true},
		{"https_proxy_protocol", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Connection", "close")
				io.WriteString(w, "complete response")
			}))
			defer upstream.Close()
			p, err := NewProxy(&Options{NewCaFunc: cert.NewSelfSignCAMemory})
			if err != nil {
				t.Fatal(err)
			}
			p.SetUpstreamProxy(func(*http.Request) (*url.URL, error) { return nil, nil })
			p.SetAuthProxy(func(_ http.ResponseWriter, r *http.Request) (bool, error) {
				return r.Header.Get("Proxy-Authorization") == "Basic dG9rZW46dGVzdA==", nil
			})
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			address := listener.Addr().String()
			defer listener.Close()
			if tc.pp {
				listener = proxyprotocol.NewListener(listener, 0)
			}
			roots := x509.NewCertPool()
			root := p.GetCertificate()
			roots.AddCert(&root)
			if tc.tls {
				certificate, err := p.GetCertificateByCN("127.0.0.1")
				if err != nil {
					t.Fatal(err)
				}
				p.SetTlsConfig(&tls.Config{Certificates: []tls.Certificate{*certificate}, MinVersion: tls.VersionTLS12})
				listener = tls.NewListener(listener, p.tlsConfig)
			}
			done := make(chan error, 1)
			go func() { done <- p.entry.server.Serve(&wrapListener{Listener: listener, proxy: p}) }()
			defer func() {
				p.Close()
				select {
				case err := <-done:
					if err != http.ErrServerClosed {
						t.Errorf("entry: %v", err)
					}
				case <-time.After(time.Second):
					t.Error("entry did not stop")
				}
			}()
			proxyURL := &url.URL{Scheme: "http", Host: address, User: url.UserPassword("token", "test")}
			if tc.tls {
				proxyURL.Scheme = "https"
			}
			transport := &http.Transport{
				Proxy:             http.ProxyURL(proxyURL),
				TLSClientConfig:   &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12},
				DisableKeepAlives: true,
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
					if err == nil && tc.pp {
						_, err = io.WriteString(conn, "PROXY TCP4 192.0.2.1 192.0.2.2 12345 8080\r\n")
						if err != nil {
							conn.Close()
							return nil, err
						}
					}
					return conn, err
				},
			}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
			response, err := client.Get(upstream.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK || string(body) != "complete response" {
				t.Fatalf("response: status=%d body=%q", response.StatusCode, body)
			}
		})
	}
}
