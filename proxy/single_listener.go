package proxy

import (
	"context"
	"net"
	"net/http"
)

type singleConnListener struct{ c net.Conn }

func (l *singleConnListener) Accept() (net.Conn, error) {
	if l.c == nil {
		return nil, net.ErrClosed
	}
	c := l.c
	l.c = nil
	return c, nil
}
func (l *singleConnListener) Close() error   { return nil }
func (l *singleConnListener) Addr() net.Addr { return l.c.LocalAddr() }

func interceptConnectHTTP(res http.ResponseWriter, req *http.Request, e *entry) error {
	// 1) 回 200
	res.WriteHeader(http.StatusOK)
	clientConn, _, err := res.(http.Hijacker).Hijack()
	if err != nil {
		return err
	}

	// 2) 取出并注入 ConnContext
	ctx := req.Context()
	cc := ctx.Value(connContextKey).(*ConnContext)

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), connContextKey, cc))
			e.ServeHTTP(w, r)
		}),
		//ErrorLog: log.New(io.Discard, "", 0),
	}

	return srv.Serve(&singleConnListener{c: clientConn})
}
