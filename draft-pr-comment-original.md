Thanks for the thorough RFC, @maxdubrinsky. The transport/auth extraction from
the CLI into a dedicated crate makes sense regardless of the FFI question, and
the decision table is well-researched.

That said, I'd like to challenge the core premise: **is the shared logic
substantial enough to justify the FFI layer?**

### What would actually be shared

I looked at the code that would move into `openshell-sdk`:

| Module | Lines | What it does |
|--------|-------|-------------|
| `tls.rs` | 451 | TLS config, channel construction, cert loading |
| `oidc_auth.rs` | 534 | OIDC browser flow, token refresh |
| `edge_tunnel.rs` | 245 | Cloudflare Access tunnel proxy |
| **Total** | **~1,230** | **Transport/auth plumbing** |

The SDK methods proposed for the MVP (sandbox CRUD, health, exec, wait) are thin
wrappers over the proto API. For example, the Python SDK's `wait_ready` is a
10-line poll loop; `Create`/`Get`/`Delete` are 5-line proto call + type
conversion each. The entire Python SDK including transport setup is 1,382 lines.
None of this is algorithmic complexity that benefits from a single
implementation -- it's configuration assembly that every language's gRPC stack
handles natively.

### Where Rust-core-with-FFI works (and why this case is different)

The prior art cited in the RFC is solid, but each case shares a key trait that
OpenShell doesn't have yet:

- **[Polars](https://github.com/pola-rs/polars)**: thousands of optimized query
  operators where correctness and performance depend on a single implementation.
- **[Temporal](https://www.infoq.com/news/2025/11/temporal-rust-polygot-sdk/)**:
  complex durable-execution state machines that *must* behave identically across
  languages. Spencer Judge explicitly noted that async FFI bridging is
  "particularly challenging" even for that use case.
- **[Statsig](https://www.statsig.com/blog/escaping-sdk-maintenance-hell)**:
  24 SDKs, 7 developers, complex feature-flag evaluation engine. Even they
  acknowledge the approach "makes things a little worse before they get better."

The common thread: **substantial, correctness-critical business logic** where
behavioral divergence between languages is a bug, not just an inconvenience.
OpenShell's proposed shared core is transport configuration and auth token
management -- well-understood problems that every gRPC language ecosystem already
solves with battle-tested libraries.

### Proto versioning is the shared contract

One argument for the Rust core is that it prevents SDK consumers from getting out
of sync when the API evolves. But the proto definition already serves this role:
when you add a field to `CreateSandboxRequest`, every language's protoc codegen
picks it up. When you add a new RPC, the SDK in each language still needs a new
wrapper method, whether that wrapper calls the proto stub directly or calls into
the Rust core through napi-rs/PyO3/cgo. The number of touchpoints doesn't
shrink.

Kubernetes faces the same problem at larger scale: 10+ officially supported
client libraries, a rapidly evolving API surface, and strict behavioral parity
requirements. Their solution is
[generating independent native clients from the OpenAPI spec](https://kubernetes.io/docs/reference/using-api/client-libraries/) --
no shared runtime core. The spec itself is the single source of truth.

### The real costs

The RFC itself flags "napi-rs prebuilt binary CI complexity" as a risk and notes
the six-target build matrix "has only been exercised on darwin-arm64 so far."
That's not a one-time setup cost. Every Rust edition bump, every GitHub runner
architecture change, every new platform target reopens that work. It shifts
ongoing effort from "write gRPC client code" (well-understood, per-language) to
"maintain cross-platform native binary CI" (specialized, fragile). It also raises the contributor barrier: every SDK
contributor needs Rust *and* the FFI binding layer *and* the target language.

For Go specifically, cgo brings additional pain: it
[disables cross-compilation by default](https://dave.cheney.net/2016/01/18/cgo-is-not-go),
breaks static linking, slows builds, and loses Go's goroutine scheduling
advantages. The Go ecosystem treats cgo as a last resort, not a default. Users
expect `go get` to just work without a C/Rust toolchain installed. A Go SDK
built on a Rust core via cgo would be non-idiomatic and harder to adopt than a
pure-Go implementation.

### Proposal: start native, share tests

Instead of investing in the FFI layer now, I'd suggest:

1. **Extract `openshell-sdk` as a Rust crate** (Phase 1 of the RFC, which makes
   sense on its own -- clean up the CLI).
2. **Build native TypeScript and Go SDKs** using each language's gRPC ecosystem.
   The SDK methods are thin enough that the per-language effort is modest.
3. **Share a conformance test suite** against a common test gateway that verifies
   behavioral parity across SDKs. This catches the "forgot to update the Python
   SDK" problem directly and is independent of implementation language.
4. **Revisit the shared-core question** once the SDK surface grows to include
   genuinely complex client-side logic (local caching, sophisticated
   retry/circuit-breaking, client-side policy evaluation) where a single
   implementation truly pays off.

This approach ships TypeScript and Go SDKs faster, keeps each SDK idiomatic to
its ecosystem, avoids the FFI/CI overhead, and leaves the door open for a shared
core when the complexity warrants it.

Happy to help flesh out any of these points or contribute to the conformance test
approach.
