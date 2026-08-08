# Proto Gap Fixes: Drop D Network Policy Types

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add domain types, converters, and tests for three sets of upstream proto fields: SigV4/JSON-RPC on NetworkEndpoint (#37), MCP policy types (#35), and network middleware types (#36).

**Architecture:** Follow the established 3-layer pattern: domain types in `types/`, converter functions in `internal/converter/`, coverage test entries in `coverage_test.go`. Each proto field gets a corresponding Go struct field, a FromProto conversion, a ToProto conversion, and round-trip/deep-copy tests.

**Tech Stack:** Go, protobuf, testify (assert/require)

## Global Constraints

- Every `.go` file must start with the SPDX license header
- All slice/map fields must be deep-copied at the proto/SDK boundary
- `*bool` for proto `optional bool` fields, using `CopyBoolPtr()` helper
- Coverage test must list every proto field in either `handled` or `skipped`
- `mise run ci` must pass after each task

---

### Task 1: SigV4 and JSON-RPC Fields on NetworkEndpoint (#37)

**Files:**
- Modify: `openshell/v1/types/network_policy.go:19-38` (PolicyNetworkEndpoint struct)
- Modify: `openshell/v1/internal/converter/network_policy.go:65-113` (policyNetworkEndpointFromProto)
- Modify: `openshell/v1/internal/converter/network_policy.go:115-157` (policyNetworkEndpointToProto)
- Modify: `openshell/v1/internal/converter/network_policy_test.go`

**Interfaces:**
- Consumes: `sbv1.NetworkEndpoint` getters: `GetCredentialSigning()`, `GetSigningService()`, `GetSigningRegion()`, `GetJsonRpcMaxBodyBytes()`
- Produces: Four new fields on `types.PolicyNetworkEndpoint`: `CredentialSigning string`, `SigningService string`, `SigningRegion string`, `JsonRpcMaxBodyBytes uint32`

- [ ] **Step 1: Add fields to PolicyNetworkEndpoint**

In `openshell/v1/types/network_policy.go`, add four fields after `AdvisorProposed`:

```go
AdvisorProposed             bool
CredentialSigning           string
SigningService              string
SigningRegion               string
JsonRpcMaxBodyBytes         uint32
```

- [ ] **Step 2: Add FromProto conversion**

In `openshell/v1/internal/converter/network_policy.go`, in `policyNetworkEndpointFromProto`, add after the `AdvisorProposed` line:

```go
CredentialSigning:   ep.GetCredentialSigning(),
SigningService:      ep.GetSigningService(),
SigningRegion:       ep.GetSigningRegion(),
JsonRpcMaxBodyBytes: ep.GetJsonRpcMaxBodyBytes(),
```

- [ ] **Step 3: Add ToProto conversion**

In `policyNetworkEndpointToProto`, add after `AdvisorProposed`:

```go
CredentialSigning:   ep.CredentialSigning,
SigningService:      ep.SigningService,
SigningRegion:       ep.SigningRegion,
JsonRpcMaxBodyBytes: ep.JsonRpcMaxBodyBytes,
```

- [ ] **Step 4: Update TestNetworkPolicyRuleFromProto**

In `network_policy_test.go`, add to the proto endpoint literal:

```go
CredentialSigning:   "sigv4",
SigningService:      "bedrock",
SigningRegion:       "us-west-2",
JsonRpcMaxBodyBytes: 65536,
```

Add assertions after the `AdvisorProposed` assertion:

```go
assert.Equal(t, "sigv4", ep.CredentialSigning)
assert.Equal(t, "bedrock", ep.SigningService)
assert.Equal(t, "us-west-2", ep.SigningRegion)
assert.Equal(t, uint32(65536), ep.JsonRpcMaxBodyBytes)
```

- [ ] **Step 5: Update TestNetworkPolicyRuleRoundTrip**

Add the four fields to the SDK endpoint literal in the round-trip test:

```go
CredentialSigning:   "sigv4",
SigningService:      "bedrock",
SigningRegion:       "us-east-1",
JsonRpcMaxBodyBytes: 32768,
```

Add assertions:

```go
assert.Equal(t, original.Endpoints[0].CredentialSigning, roundTrip.Endpoints[0].CredentialSigning)
assert.Equal(t, original.Endpoints[0].SigningService, roundTrip.Endpoints[0].SigningService)
assert.Equal(t, original.Endpoints[0].SigningRegion, roundTrip.Endpoints[0].SigningRegion)
assert.Equal(t, original.Endpoints[0].JsonRpcMaxBodyBytes, roundTrip.Endpoints[0].JsonRpcMaxBodyBytes)
```

- [ ] **Step 6: Run tests**

Run: `mise run test`
Expected: All tests pass including coverage test (fields already in `handled` set).

- [ ] **Step 7: Commit**

```bash
git add openshell/v1/types/network_policy.go openshell/v1/internal/converter/network_policy.go openshell/v1/internal/converter/network_policy_test.go
git commit -m "feat: add SigV4 and JSON-RPC fields to PolicyNetworkEndpoint (#37)"
```

---

### Task 2: MCP Policy Types (#35)

**Files:**
- Modify: `openshell/v1/types/network_policy.go` (add McpOptions type, Mcp field on endpoint, Params on L7Allow/L7DenyRule)
- Modify: `openshell/v1/internal/converter/network_policy.go` (McpOptions converter, Params converter, wire into endpoint/L7 converters)
- Modify: `openshell/v1/internal/converter/network_policy_test.go`

**Interfaces:**
- Consumes: `sbv1.McpOptions` with `GetStrictToolNames() *bool`, `GetAllowAllKnownMcpMethods() *bool`; `sbv1.NetworkEndpoint.GetMcp() *sbv1.McpOptions`; `sbv1.L7Allow.GetParams()`, `sbv1.L7DenyRule.GetParams()` returning `map[string]*sbv1.L7QueryMatcher`
- Produces: `types.McpOptions` struct; `types.PolicyNetworkEndpoint.Mcp *McpOptions`; `types.L7Allow.Params map[string]L7QueryMatcher`; `types.L7DenyRule.Params map[string]L7QueryMatcher`

- [ ] **Step 1: Add McpOptions domain type**

In `openshell/v1/types/network_policy.go`, add after the `GraphqlOperation` struct:

```go
// McpOptions configures MCP-specific policy controls on a network endpoint.
type McpOptions struct {
	StrictToolNames         *bool
	AllowAllKnownMcpMethods *bool
}
```

- [ ] **Step 2: Add Mcp field to PolicyNetworkEndpoint**

In the `PolicyNetworkEndpoint` struct, add after `JsonRpcMaxBodyBytes`:

```go
Mcp *McpOptions
```

- [ ] **Step 3: Add Params field to L7Allow and L7DenyRule**

In the `L7Allow` struct, add after `Fields`:

```go
Params map[string]L7QueryMatcher
```

In the `L7DenyRule` struct, add after `Fields`:

```go
Params map[string]L7QueryMatcher
```

- [ ] **Step 4: Add McpOptions converter functions**

In `openshell/v1/internal/converter/network_policy.go`, add:

```go
func mcpOptionsFromProto(m *sbv1.McpOptions) *types.McpOptions {
	if m == nil {
		return nil
	}
	return &types.McpOptions{
		StrictToolNames:         CopyBoolPtr(m.StrictToolNames),
		AllowAllKnownMcpMethods: CopyBoolPtr(m.AllowAllKnownMcpMethods),
	}
}

func mcpOptionsToProto(m *types.McpOptions) *sbv1.McpOptions {
	if m == nil {
		return nil
	}
	return &sbv1.McpOptions{
		StrictToolNames:         CopyBoolPtr(m.StrictToolNames),
		AllowAllKnownMcpMethods: CopyBoolPtr(m.AllowAllKnownMcpMethods),
	}
}
```

- [ ] **Step 5: Wire Mcp into endpoint converters**

In `policyNetworkEndpointFromProto`, add after the struct literal (near the GraphQL block):

```go
result.Mcp = mcpOptionsFromProto(ep.GetMcp())
```

In `policyNetworkEndpointToProto`, add after the GraphQL block:

```go
result.Mcp = mcpOptionsToProto(ep.Mcp)
```

- [ ] **Step 6: Wire Params into L7Allow converter**

In `l7RuleFromProto`, inside the `if a := r.GetAllow(); a != nil` block, add after the Query handling:

```go
if p := a.GetParams(); len(p) > 0 {
	result.Allow.Params = l7QueryMapFromProto(p)
}
```

In `l7RuleToProto`, inside the `if r.Allow != nil` block, add after the Query handling:

```go
if len(r.Allow.Params) > 0 {
	result.Allow.Params = l7QueryMapToProto(r.Allow.Params)
}
```

- [ ] **Step 7: Wire Params into L7DenyRule converter**

In `l7DenyRuleFromProto`, add `Params` to the return struct:

```go
Params: l7QueryMapFromProtoDeny(r.GetParams()),
```

In `l7DenyRuleToProto`, add after the Query handling:

```go
if len(r.Params) > 0 {
	result.Params = l7QueryMapToProtoDeny(r.Params)
}
```

- [ ] **Step 8: Add MCP tests**

In `network_policy_test.go`, update `TestNetworkPolicyRuleFromProto` proto literal to include MCP:

```go
Mcp: &sbv1.McpOptions{
	StrictToolNames:         boolPtr(true),
	AllowAllKnownMcpMethods: boolPtr(false),
},
```

Add a `boolPtr` helper at the bottom of the test file:

```go
func boolPtr(v bool) *bool { return &v }
```

Add assertions:

```go
require.NotNil(t, ep.Mcp)
require.NotNil(t, ep.Mcp.StrictToolNames)
assert.True(t, *ep.Mcp.StrictToolNames)
require.NotNil(t, ep.Mcp.AllowAllKnownMcpMethods)
assert.False(t, *ep.Mcp.AllowAllKnownMcpMethods)
```

Also add `Params` to the L7Allow proto in the test:

```go
Params: map[string]*sbv1.L7QueryMatcher{
	"name": {Glob: "my-tool-*"},
},
```

And assert:

```go
require.Contains(t, allow.Params, "name")
assert.Equal(t, "my-tool-*", allow.Params["name"].Glob)
```

Add `Params` to the L7DenyRule proto and assert similarly.

- [ ] **Step 9: Add MCP round-trip test data**

Update `TestNetworkPolicyRuleRoundTrip` SDK literal to include:

```go
Mcp: &v1.McpOptions{
	StrictToolNames:         boolPtr(true),
	AllowAllKnownMcpMethods: boolPtr(false),
},
```

And add Params to the L7Allow and L7DenyRule in the round-trip test.

Add assertions for MCP round-trip:

```go
require.NotNil(t, roundTrip.Endpoints[0].Mcp)
assert.Equal(t, original.Endpoints[0].Mcp.StrictToolNames, roundTrip.Endpoints[0].Mcp.StrictToolNames)
```

- [ ] **Step 10: Add MCP deep-copy test**

Add to `TestNetworkPolicyRuleDeepCopy`:

```go
// Also test MCP deep copy
mcpProto := &sbv1.NetworkPolicyRule{
	Name: "mcp-test",
	Endpoints: []*sbv1.NetworkEndpoint{
		{
			Mcp: &sbv1.McpOptions{
				StrictToolNames: boolPtr(true),
			},
		},
	},
}
mcpRule := NetworkPolicyRuleFromProto(mcpProto)
*mcpProto.Endpoints[0].Mcp.StrictToolNames = false
require.NotNil(t, mcpRule.Endpoints[0].Mcp.StrictToolNames)
assert.True(t, *mcpRule.Endpoints[0].Mcp.StrictToolNames)
```

- [ ] **Step 11: Run tests**

Run: `mise run test`
Expected: All tests pass.

- [ ] **Step 12: Commit**

```bash
git add openshell/v1/types/network_policy.go openshell/v1/internal/converter/network_policy.go openshell/v1/internal/converter/network_policy_test.go
git commit -m "feat: add MCP policy types and Params field (#35)"
```

---

### Task 3: Network Middleware Types (#36)

**Files:**
- Modify: `openshell/v1/types/policy.go` (add 3 new types, NetworkMiddlewares field on SandboxPolicy)
- Modify: `openshell/v1/internal/converter/policy.go` (middleware converter functions, wire into SandboxPolicy converter)
- Modify: `openshell/v1/internal/converter/coverage_test.go` (move network_middlewares from skipped to handled)
- Modify: `openshell/v1/internal/converter/policy_test.go` (tests for middleware conversion)

**Interfaces:**
- Consumes: `sbv1.NetworkMiddlewareConfig` with getters: `GetName()`, `GetMiddleware()`, `GetConfig() *structpb.Struct`, `GetOnError()`, `GetEndpoints() *sbv1.MiddlewareEndpointSelector`, `GetOrder() int32`; `sbv1.MiddlewareEndpointSelector` with `GetInclude()`, `GetExclude()`; `sbv1.SandboxPolicy.GetNetworkMiddlewares() map[string]*sbv1.NetworkMiddlewareConfig`
- Produces: `types.NetworkMiddlewareConfig`, `types.MiddlewareEndpointSelector`, `types.SandboxPolicy.NetworkMiddlewares map[string]NetworkMiddlewareConfig`

Note: `SupervisorMiddlewareService` is on `GetSandboxConfigResponse`, not `SandboxPolicy`. It belongs to a config client converter, not the policy converter. Leave it out of this task (it will be handled when the config client is implemented in Drop D).

- [ ] **Step 1: Add middleware domain types**

In `openshell/v1/types/policy.go`, add after the `SandboxPolicy` struct (before `FilesystemPolicy`):

```go
// NetworkMiddlewareConfig configures a supervisor middleware pipeline for
// network egress. Middleware configs are referenced by name in the policy.
type NetworkMiddlewareConfig struct {
	Name       string
	Middleware  string
	Config     map[string]any
	OnError    string
	Endpoints  *MiddlewareEndpointSelector
	Order      int32
}

// MiddlewareEndpointSelector controls which admitted destinations use a
// middleware config, using host glob patterns.
type MiddlewareEndpointSelector struct {
	Include []string
	Exclude []string
}
```

- [ ] **Step 2: Add NetworkMiddlewares field to SandboxPolicy**

In the `SandboxPolicy` struct, add after `NetworkPolicies`:

```go
NetworkMiddlewares map[string]NetworkMiddlewareConfig
```

- [ ] **Step 3: Add middleware converter functions**

In `openshell/v1/internal/converter/policy.go`, add the import for `structpb`:

```go
import (
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	sbv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
	"google.golang.org/protobuf/types/known/structpb"
)
```

Add converter functions:

```go
func middlewareConfigFromProto(m *sbv1.NetworkMiddlewareConfig) types.NetworkMiddlewareConfig {
	result := types.NetworkMiddlewareConfig{
		Name:       m.GetName(),
		Middleware: m.GetMiddleware(),
		OnError:    m.GetOnError(),
		Order:      m.GetOrder(),
	}
	if c := m.GetConfig(); c != nil {
		result.Config = c.AsMap()
	}
	if ep := m.GetEndpoints(); ep != nil {
		result.Endpoints = &types.MiddlewareEndpointSelector{
			Include: CopyStringSlice(ep.GetInclude()),
			Exclude: CopyStringSlice(ep.GetExclude()),
		}
	}
	return result
}

func middlewareConfigToProto(m *types.NetworkMiddlewareConfig) *sbv1.NetworkMiddlewareConfig {
	result := &sbv1.NetworkMiddlewareConfig{
		Name:       m.Name,
		Middleware: m.Middleware,
		OnError:    m.OnError,
		Order:      m.Order,
	}
	if m.Config != nil {
		s, err := structpb.NewStruct(m.Config)
		if err == nil {
			result.Config = s
		}
	}
	if m.Endpoints != nil {
		result.Endpoints = &sbv1.MiddlewareEndpointSelector{
			Include: CopyStringSlice(m.Endpoints.Include),
			Exclude: CopyStringSlice(m.Endpoints.Exclude),
		}
	}
	return result
}
```

- [ ] **Step 4: Wire middleware into SandboxPolicy converters**

In `SandboxPolicyFromProto`, add after the `NetworkPolicies` block:

```go
if mw := p.GetNetworkMiddlewares(); mw != nil {
	result.NetworkMiddlewares = make(map[string]types.NetworkMiddlewareConfig, len(mw))
	for k, v := range mw {
		if v != nil {
			result.NetworkMiddlewares[k] = middlewareConfigFromProto(v)
		}
	}
}
```

In `SandboxPolicyToProto`, add after the `NetworkPolicies` block:

```go
if p.NetworkMiddlewares != nil {
	result.NetworkMiddlewares = make(map[string]*sbv1.NetworkMiddlewareConfig, len(p.NetworkMiddlewares))
	for k, v := range p.NetworkMiddlewares {
		result.NetworkMiddlewares[k] = middlewareConfigToProto(&v)
	}
}
```

- [ ] **Step 5: Update coverage test**

In `coverage_test.go`, in `TestConverterCoversAllProtoFields_SandboxPolicy`:

Move `"network_middlewares"` from `skipped` to `handled`:

```go
handled := fieldSet{
	"version":             true,
	"filesystem":          true,
	"network_policies":    true,
	"process":             true,
	"landlock":            true,
	"network_middlewares": true,
}
```

Remove the `skipped` variable and pass `nil` for skipped:

```go
assertAllFieldsCovered(t, (&sandboxpb.SandboxPolicy{}).ProtoReflect().Descriptor(), handled, nil)
```

Add a new coverage test for `NetworkMiddlewareConfig`:

```go
func TestConverterCoversAllProtoFields_NetworkMiddlewareConfig(t *testing.T) {
	handled := fieldSet{
		"name":       true,
		"middleware": true,
		"config":     true,
		"on_error":   true,
		"endpoints":  true,
		"order":      true,
	}

	assertAllFieldsCovered(t, (&sandboxpb.NetworkMiddlewareConfig{}).ProtoReflect().Descriptor(), handled, nil)
}

func TestConverterCoversAllProtoFields_MiddlewareEndpointSelector(t *testing.T) {
	handled := fieldSet{
		"include": true,
		"exclude": true,
	}

	assertAllFieldsCovered(t, (&sandboxpb.MiddlewareEndpointSelector{}).ProtoReflect().Descriptor(), handled, nil)
}
```

- [ ] **Step 6: Add middleware converter tests**

In `openshell/v1/internal/converter/policy_test.go`, add:

```go
func TestSandboxPolicyFromProto_WithMiddleware(t *testing.T) {
	proto := &sbv1.SandboxPolicy{
		Version: 3,
		NetworkMiddlewares: map[string]*sbv1.NetworkMiddlewareConfig{
			"sigv4-rewriter": {
				Name:       "sigv4-rewriter",
				Middleware: "aws-sigv4",
				OnError:    "fail_closed",
				Order:      10,
				Config: func() *structpb.Struct {
					s, _ := structpb.NewStruct(map[string]any{
						"region":  "us-east-1",
						"service": "bedrock",
					})
					return s
				}(),
				Endpoints: &sbv1.MiddlewareEndpointSelector{
					Include: []string{"*.bedrock.amazonaws.com"},
					Exclude: []string{"sts.amazonaws.com"},
				},
			},
		},
	}

	policy := SandboxPolicyFromProto(proto)

	require.NotNil(t, policy)
	require.Contains(t, policy.NetworkMiddlewares, "sigv4-rewriter")
	mw := policy.NetworkMiddlewares["sigv4-rewriter"]
	assert.Equal(t, "sigv4-rewriter", mw.Name)
	assert.Equal(t, "aws-sigv4", mw.Middleware)
	assert.Equal(t, "fail_closed", mw.OnError)
	assert.Equal(t, int32(10), mw.Order)
	require.NotNil(t, mw.Config)
	assert.Equal(t, "us-east-1", mw.Config["region"])
	assert.Equal(t, "bedrock", mw.Config["service"])
	require.NotNil(t, mw.Endpoints)
	assert.Equal(t, []string{"*.bedrock.amazonaws.com"}, mw.Endpoints.Include)
	assert.Equal(t, []string{"sts.amazonaws.com"}, mw.Endpoints.Exclude)
}

func TestSandboxPolicyMiddlewareRoundTrip(t *testing.T) {
	original := &types.SandboxPolicy{
		Version: 5,
		NetworkMiddlewares: map[string]types.NetworkMiddlewareConfig{
			"rate-limiter": {
				Name:       "rate-limiter",
				Middleware: "envoy-ratelimit",
				OnError:    "fail_open",
				Order:      20,
				Config: map[string]any{
					"requests_per_second": float64(100),
				},
				Endpoints: &types.MiddlewareEndpointSelector{
					Include: []string{"api.*"},
				},
			},
		},
	}

	proto := SandboxPolicyToProto(original)
	require.NotNil(t, proto)

	roundTrip := SandboxPolicyFromProto(proto)
	require.NotNil(t, roundTrip)

	require.Contains(t, roundTrip.NetworkMiddlewares, "rate-limiter")
	mw := roundTrip.NetworkMiddlewares["rate-limiter"]
	assert.Equal(t, original.NetworkMiddlewares["rate-limiter"].Name, mw.Name)
	assert.Equal(t, original.NetworkMiddlewares["rate-limiter"].Middleware, mw.Middleware)
	assert.Equal(t, original.NetworkMiddlewares["rate-limiter"].OnError, mw.OnError)
	assert.Equal(t, original.NetworkMiddlewares["rate-limiter"].Order, mw.Order)
	assert.Equal(t, original.NetworkMiddlewares["rate-limiter"].Config["requests_per_second"], mw.Config["requests_per_second"])
	assert.Equal(t, original.NetworkMiddlewares["rate-limiter"].Endpoints.Include, mw.Endpoints.Include)
}

func TestSandboxPolicyMiddlewareDeepCopy(t *testing.T) {
	proto := &sbv1.SandboxPolicy{
		NetworkMiddlewares: map[string]*sbv1.NetworkMiddlewareConfig{
			"test": {
				Endpoints: &sbv1.MiddlewareEndpointSelector{
					Include: []string{"original.com"},
				},
			},
		},
	}

	policy := SandboxPolicyFromProto(proto)
	proto.NetworkMiddlewares["test"].Endpoints.Include[0] = "mutated.com"

	assert.Equal(t, "original.com", policy.NetworkMiddlewares["test"].Endpoints.Include[0])
}
```

Add the `structpb` import to the test file:

```go
import (
	"google.golang.org/protobuf/types/known/structpb"
)
```

- [ ] **Step 7: Run tests**

Run: `mise run test`
Expected: All tests pass, coverage test has no skipped fields for SandboxPolicy.

- [ ] **Step 8: Commit**

```bash
git add openshell/v1/types/policy.go openshell/v1/internal/converter/policy.go openshell/v1/internal/converter/coverage_test.go openshell/v1/internal/converter/policy_test.go
git commit -m "feat: add network middleware types and SandboxPolicy.NetworkMiddlewares (#36)"
```

---

### Task 4: Final Verification and CI

**Files:** None (verification only)

**Interfaces:**
- Consumes: All changes from Tasks 1-3
- Produces: Green CI

- [ ] **Step 1: Run full CI**

Run: `mise run ci`
Expected: lint, build, test, proto:check all pass.

- [ ] **Step 2: Verify no coverage test gaps**

Run: `go test -run TestConverterCoversAllProtoFields ./openshell/v1/internal/converter/ -v`
Expected: All coverage tests pass, no "not handled" or "not explicitly skipped" errors.

- [ ] **Step 3: Verify proto:check**

Run: `mise run proto:check`
Expected: Generated files are up to date (proto was regenerated at the start).
