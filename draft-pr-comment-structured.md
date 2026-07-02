Nice RFC, @maxdubrinsky. Extracting the transport/auth code from the CLI into its own crate is a good idea regardless of where the FFI discussion lands.

I've been looking at this from the perspective of a potential Go SDK and wanted to raise a question: is ~1,230 lines of transport plumbing enough shared logic to justify the FFI layer?

### What would actually be shared

I went through the code that would move into `openshell-sdk`:

| Module | Lines | What it does |
|--------|-------|-------------|
| `tls.rs` | 451 | TLS config, channel construction, cert loading |
| `oidc_auth.rs` | 534 | OIDC browser flow, token refresh |
| `edge_tunnel.rs` | 245 | Cloudflare Access tunnel proxy |

The SDK methods proposed for the MVP (sandbox CRUD, health, exec, wait) are thin wrappers over the proto API. The Python SDK's `wait_ready` is a 10-line poll loop. `Create`/`Get`/`Delete` are 5-line proto call + type conversion each. The whole Python SDK including transport setup is 1,382 lines. This isn't algorithmic complexity where a single implementation prevents bugs. It's configuration assembly, and every language's gRPC stack already handles it natively.

### The prior art is compelling, but the scale is different

The RFC cites [Polars](https://github.com/pola-rs/polars), [Temporal](https://www.infoq.com/news/2025/11/temporal-rust-polygot-sdk/), and [Statsig](https://www.statsig.com/blog/escaping-sdk-maintenance-hell). All solid examples. I wonder, though, whether they share a trait that OpenShell doesn't have yet: massive, correctness-critical business logic where behavioral divergence between languages would be an actual bug.

Polars has thousands of optimized query operators. Temporal has durable-execution state machines that must behave identically across languages (and Spencer Judge still called async FFI bridging "particularly challenging"). Statsig runs a complex feature-flag evaluation engine across 24 SDKs with 7 developers, and they're candid that the approach "makes things a little worse before they get better."

OpenShell's proposed shared core is transport configuration and auth token management. Those are well-understood problems with mature, battle-tested libraries in every gRPC ecosystem.

### Proto versioning already prevents SDK drift

One argument for the Rust core is preventing SDK consumers from getting out of sync when the API evolves. But the proto definition already does this: when you add a field to `CreateSandboxRequest`, every language's protoc codegen picks it up. When you add a new RPC, the SDK in each language still needs a new wrapper method, whether that wrapper calls the proto stub directly or goes through napi-rs/PyO3/cgo. The number of touchpoints stays the same.

Kubernetes deals with this at larger scale (10+ officially supported client libraries, rapidly evolving API) and their approach is [generating independent native clients from the OpenAPI spec](https://kubernetes.io/docs/reference/using-api/client-libraries/). No shared runtime core. The spec itself is the single source of truth.

### Costs worth considering

The RFC flags "napi-rs prebuilt binary CI complexity" as a risk and notes the six-target build matrix "has only been exercised on darwin-arm64 so far." That's not a one-time setup cost. Every Rust edition bump, every GitHub runner architecture change, every new platform target reopens that work. You're shifting effort from "write gRPC client code" (well-understood) to "maintain cross-platform native binary CI" (specialized, fragile). And every SDK contributor now needs Rust plus the binding layer plus the target language.

For Go, cgo makes this worse. It [disables cross-compilation by default](https://dave.cheney.net/2016/01/18/cgo-is-not-go), breaks static linking, slows builds, and loses Go's goroutine scheduling. The Go ecosystem treats cgo as a last resort. Users expect `go get` to work without a C/Rust toolchain installed. A Go SDK on a Rust core via cgo would be harder to adopt than a pure-Go one.

### Alternative worth exploring: start native, share tests

What if we kept Phase 1 of the RFC and built native SDKs instead? Concretely:

1. **Extract `openshell-sdk` as a Rust crate** (Phase 1 of the RFC, which makes sense on its own to clean up the CLI).
2. **Build native TypeScript and Go SDKs** using each language's gRPC ecosystem. The SDK methods are thin enough that the per-language effort is modest.
3. **Share a conformance test suite** against a common test gateway that verifies behavioral parity across SDKs. This catches the "forgot to update the Python SDK" problem directly and is independent of implementation language.
4. **Revisit the shared-core question** once the SDK surface grows to include genuinely complex client-side logic (local caching, retry with circuit-breaking, client-side policy evaluation). That would be the point where a shared core clearly pays for itself, and it could be revisited then.

This approach ships TypeScript and Go SDKs faster, keeps each SDK idiomatic to its ecosystem, avoids the FFI/CI overhead, and leaves the door open for a shared core when the complexity warrants it.

I'm new to the project but happy to dig into any of these points or help shape the conformance test approach if there's interest.
