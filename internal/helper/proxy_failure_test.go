package helper

import (
	"context"
	"net"
	"net/url"
	"testing"
	"time"
)

func TestGetProxyConnSocks5DialFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	proxyURL := &url.URL{Scheme: "socks5", Host: address}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := GetProxyConn(ctx, proxyURL, "127.0.0.1:80", false)
	if conn != nil {
		conn.Close()
		t.Fatal("failed SOCKS5 dial returned a connection")
	}
	if err == nil {
		t.Fatal("expected SOCKS5 dial failure")
	}
}
