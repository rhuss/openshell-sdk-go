Small correction to my Go SDK numbers above: the transport/auth layer is actually **220 lines** (not 143), once you include the internal TLS credential builder. It covers all four TLS modes from the RFC (plaintext, CA-only, mTLS, insecure) plus static bearer token auth.

What it doesn't include: OIDC interactive flows (browser auth, token refresh, discovery) and the Cloudflare Access tunnel. But looking at the RFC's own design, those are explicitly scoped out of `openshell-sdk` too ("SDK never sees a browser", "auth token file loading NOT in openshell-sdk directly"). The OIDC browser flow stays in the CLI, and the SDK consumes a `Refresh` trait that the CLI implements.

So the 220 lines in Go and the ~1,230 lines in Rust aren't really comparable at face value. Roughly half of the Rust OIDC code (lines 300-534 of `oidc_auth.rs`) is a localhost HTTP callback server for the browser flow plus tests. That's CLI-specific UX, not SDK logic that would be shared through FFI bindings.

The part of the Rust transport code that would actually cross the FFI boundary into language bindings is closer to the same scope the Go SDK already covers natively in 220 lines.

I'm planning to implement the remaining pieces (OIDC token refresh, single-flight coalescing, `from_active_cluster` config loading) in the Go prototype and will report back with the full numbers. That should give us a concrete data point on what the complete transport/auth layer looks like in a native SDK.
