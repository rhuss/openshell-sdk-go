//go:build ignore

// Contract: Public API surface for edge auth feature.
// This file documents the exported symbols and their signatures.
// It is NOT compiled; it is a design reference for implementation.

package contracts

// === Core SDK (openshell/v1/) ===

// WithExtraHeaders wraps base with additional per-RPC headers.
// Keys are normalized to lowercase. Empty-string values are silently dropped.
// Returns error if base is nil or headers is nil/empty.
func WithExtraHeaders(base AuthProvider, headers map[string]string) (AuthProvider, error)

// === Edge Package (openshell/v1/edge/) ===

// CloudflareAccess returns an AuthProvider that adds Cloudflare Access
// headers (cf-access-jwt-assertion and CF_Authorization cookie) to RPCs.
// Returns error if edgeToken is empty.
func CloudflareAccess(baseAuth AuthProvider, edgeToken string) (AuthProvider, error)

// TunnelProxy bridges gRPC connections over a WebSocket tunnel.
// The gRPC client dials TunnelProxy.Addr() instead of the remote gateway.
type TunnelProxy struct{}

// NewTunnelProxy creates a tunnel proxy that forwards gRPC traffic
// through a WebSocket connection to gatewayURL. The edgeToken authenticates
// with the edge proxy. Returns error if gatewayURL is invalid or empty.
func NewTunnelProxy(gatewayURL, edgeToken string, opts ...TunnelOption) (*TunnelProxy, error)

// Addr returns the local address the gRPC client should dial.
func (*TunnelProxy) Addr() string

// Close drains in-flight connections (up to the configured timeout,
// default 5s) then force-closes any remaining connections.
// All goroutines are cleaned up. Safe to call multiple times.
func (*TunnelProxy) Close() error

// TunnelOption configures TunnelProxy behavior.
type TunnelOption func(*tunnelConfig)

// WithTunnelLogger sets the structured logger for tunnel events.
func WithTunnelLogger(l Logger) TunnelOption

// WithTunnelTLS sets TLS configuration for the WebSocket connection (wss://).
func WithTunnelTLS(cfg *tls.Config) TunnelOption

// WithCloseTimeout sets the maximum time Close waits for in-flight
// connections to drain before force-closing. Default is 5 seconds.
func WithCloseTimeout(d time.Duration) TunnelOption
