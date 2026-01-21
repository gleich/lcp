package apis

import (
	"context"
	"net"
	"net/http"
	"time"
)

// IPV4OnlyClient returns an *http.Client that forces all outbound connections to use IPv4.
//
// This is useful in environments where IPv6 resolution exists (AAAA records are returned)
// but IPv6 egress is unavailable or unreliable (common in some Docker/container setups).
// By dialing with "tcp4" regardless of the requested network, it avoids flaky dual-stack
// behavior and intermittent request failures.
//
// The returned client uses sensible defaults for timeouts/keepalives and enables HTTP/2
// negotiation (over TLS) when supported by the server.
func IPV4OnlyClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		ForceAttemptHTTP2: true,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}
