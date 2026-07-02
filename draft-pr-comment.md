Nice RFC, @maxdubrinsky. Extracting the transport/auth code from the CLI into its own crate is a good idea regardless of where the FFI discussion lands.

I've been looking at this from the perspective of a potential Go SDK and wanted to raise a question: is ~1,230 lines of transport plumbing enough shared logic to justify the FFI layer?

### What would actually be shared

I went through the code that would move into `openshell-sdk`:

| Module | Lines | What it does |
|--------|-------|-------------|
| `tls.rs` | 451 | TLS config, channel construction, cert loading |
| `oidc_auth.rs` | 534 | OIDC browser flow, token refresh |
| `edge_tunnel.rs` | 245 | Cloudflare Access tunnel proxy |
| **Total** | **~1,230** | **Transport/auth plumbing** |

The SDK methods proposed for the MVP (sandbox CRUD, health, exec, wait) are thin wrappers over the proto API. The Python SDK's `wait_ready` is a 10-line poll loop. `Create`/`Get`/`Delete` are 5-line proto call + type conversion each. The whole Python SDK including transport setup is 1,382 lines. This isn't algorithmic complexity where a single implementation prevents bugs. It's configuration assembly, and every language's gRPC stack already handles it natively.

The RFC mentions OIDC single-flight refresh coalescing and the Cloudflare Access tunnel as areas of non-trivial logic. But single-flight token refresh is a well-known pattern that takes ~50 lines in any language (Go's `singleflight` package, Python's `asyncio.Lock`, Node's promise caching). The CF tunnel proxy is 245 lines of Rust. File transfer/upload could add real complexity, but it hasn't been designed yet.

### The prior art is compelling, but the scale is different

The RFC cites [Polars](https://github.com/pola-rs/polars), [Temporal](https://www.infoq.com/news/2025/11/temporal-rust-polygot-sdk/), and [Statsig](https://www.statsig.com/blog/escaping-sdk-maintenance-hell). All solid examples. I wonder, though, whether they share a trait that OpenShell doesn't have yet: massive, correctness-critical business logic where behavioral divergence between languages would be an actual bug.

Polars has thousands of optimized query operators. Temporal has durable-execution state machines that must behave identically across languages (and Spencer Judge still called async FFI bridging "particularly challenging"). Statsig runs a complex feature-flag evaluation engine across 24 SDKs with 7 developers, and they're candid that the approach "makes things a little worse before they get better."

OpenShell's proposed shared core is transport configuration and auth token management. Those are well-understood problems with mature, battle-tested libraries in every gRPC ecosystem.

### Proto versioning already prevents SDK drift

One argument for the Rust core is preventing SDK consumers from getting out of sync when the API evolves. But the proto definition already serves as the compile-time contract check. When you add a field to `CreateSandboxRequest`, every language's protoc codegen picks it up. Additive field changes are wire-compatible by design; gRPC has well-established patterns for field deprecation and service versioning.

When you add a new RPC (say `PauseSandbox`), the Rust-core path still requires updating the Rust crate, updating the napi-rs binding, updating the PyO3 binding, and updating any SDK that doesn't use the core (like Go). That's the same number of touchpoints as updating each native SDK directly. You've replaced "update Python gRPC client" with "update PyO3 FFI wrapper."

Kubernetes deals with this at larger scale (10+ officially supported client libraries, rapidly evolving API) and their approach is [generating independent native clients from the OpenAPI spec](https://kubernetes.io/docs/reference/using-api/client-libraries/). No shared runtime core. The spec itself is the single source of truth.

### Costs worth considering

The RFC flags "napi-rs prebuilt binary CI complexity" as a risk and notes the six-target build matrix "has only been exercised on darwin-arm64 so far." That's not a one-time setup cost. An napi-rs SDK ships prebuilt native binaries for each platform, which means CI cross-compiles Rust to every target using platform-specific toolchains (zig linker, musl, per-target Docker images). When any piece of that chain changes (a new Rust edition, GitHub switching macOS runners from Intel to M1, a new platform target), the builds can silently break. A native TypeScript SDK using `@grpc/grpc-js` or a Go SDK using `go build` simply doesn't have this class of problem. And every SDK contributor now needs Rust plus the binding layer plus the target language.

For Go, cgo makes this worse. It [disables cross-compilation by default](https://dave.cheney.net/2016/01/18/cgo-is-not-go), breaks static linking, slows builds, and loses Go's goroutine scheduling. The Go ecosystem treats cgo as a last resort. Users expect `go get` to work without a C/Rust toolchain installed. A Go SDK on a Rust core via cgo would be harder to adopt than a pure-Go one.

### Alternative: start native, share tests

1. **Extract `openshell-sdk` as a Rust crate** (Phase 1 of the RFC, which makes sense on its own to clean up the CLI).
2. **Build native TypeScript and Go SDKs** using each language's gRPC ecosystem. The SDK methods are thin enough that the per-language effort is modest.
3. **Share a conformance test suite** against a common test gateway that verifies behavioral parity across SDKs. This catches the "forgot to update the Python SDK" problem directly and is independent of implementation language.
4. **Revisit the shared-core question** once the SDK surface grows to include genuinely complex client-side logic (local caching, retry with circuit-breaking, client-side policy evaluation). That would be the point where a shared core clearly pays for itself.

### Data point: a pure-Go prototype

As a concrete example, I've been prototyping a [pure-Go SDK](https://github.com/rhuss/openshell-sdk-go) that covers the full proto surface without any Rust dependency. Some numbers on the effort beyond vanilla gRPC wrappers:

| Area | Lines | What it adds over raw proto stubs |
|------|-------|----------------------------------|
| TCP forwarding | 376 | Bidirectional stream multiplexing over `ForwardTcp` |
| Exec (streaming + interactive) | 306 | Stream chunking, stdin piping, exit-code extraction |
| SSH session management | 152 | Session lifecycle, key handling |
| Proto-to-SDK type converters | 1,894 | Idiomatic Go types, deep copy at boundaries |
| Transport/auth setup | 220 | TLS (plaintext/CA/mTLS/insecure), bearer auth, channel construction |

Total: ~8,300 lines of non-test code (plus ~13,700 lines of tests). The transport/auth layer is 220 lines covering all four TLS modes from the RFC. OIDC interactive flows and the CF tunnel are not included, but the RFC itself keeps those out of the SDK ("SDK never sees a browser", "auth token file loading NOT in openshell-sdk directly"). The bulk of the work is type conversion and the streaming RPCs (TCP, exec), not transport plumbing. All of it uses standard `google.golang.org/grpc` and compiles with a plain `go build`.

This is early-stage, but it suggests the per-language effort for a native SDK is manageable, and the parts that need the most code (type mapping, streaming) are inherently language-specific anyway.

---

I'm new to the project but happy to dig into any of these points or help shape the conformance test approach if there's interest.

Author: Roland Huß [AIA HAb SeNc Hin R Claude Opus 4.6 v1.0](https://aiattribution.github.io/statements/AIA-HAb-SeNc-Hin-R-?model=Claude%20Opus%204.6-v1.0)
