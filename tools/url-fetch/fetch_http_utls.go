package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// newUTLSHTTP2Client は utls + HTTP/2 で HTTPS する http.Client を返す。
func newUTLSHTTP2Client() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return dialUTLSChrome(ctx, network, addr)
			},
		},
	}
}

// newUTLSHTTP1Client は utls + HTTP/1.1 強制で HTTPS する http.Client を返す。
func newUTLSHTTP1Client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialUTLSChromeHTTP1(ctx, network, addr)
			},
			ForceAttemptHTTP2: false,
		},
	}
}

// dialUTLSChromeHTTP1 は utls で Chrome 指紋かつ ALPN http/1.1 のみの接続を開く。
func dialUTLSChromeHTTP1(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	tlsCfg := &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	tlsConn := utls.UClient(conn, tlsCfg, utls.HelloCustom)
	spec, err := utls.UTLSIdToSpec(utls.HelloChrome_Auto)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	for _, ext := range spec.Extensions {
		if alpn, ok := ext.(*utls.ALPNExtension); ok {
			alpn.AlpnProtocols = []string{"http/1.1"}
			break
		}
	}
	if err := tlsConn.ApplyPreset(&spec); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return tlsConn, nil
}

// dialUTLSChrome は utls で Chrome 互換 TLS ハンドシェイクを行う接続を開く。
func dialUTLSChrome(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	tlsConn := utls.UClient(conn, &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}, utls.HelloChrome_Auto)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return tlsConn, nil
}
