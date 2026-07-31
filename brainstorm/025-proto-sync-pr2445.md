# Brainstorm: Proto Sync from Upstream (PR #2445)

**Date:** 2026-07-31
**Status:** active

## Problem Framing

The SDK's proto files are ~400 lines behind upstream OpenShell after PR [#2445](https://github.com/NVIDIA/OpenShell/pull/2445) ("Wire authorization into workspace model") merged on 2026-07-30. The SDK is also missing `inference.proto` entirely. This proto sync is the prerequisite for all workspace scoping work (issues #32, #33), GatewayInfo (#34), and the new inference client.

### What PR #2445 Changed

- `openshell.proto`: +458/-64 lines. Added `AuthorizationRule` annotations to every RPC, workspace scoping fields on request messages.
- `inference.proto`: +26/-5 lines. Added auth annotations to all 4 inference RPCs.
- `options.proto`: +21 lines. Added `AuthorizationRule` message and `MethodOptions` extension (ext 50000).

### Current SDK Proto State

| File | Upstream (lines) | SDK (lines) | Delta |
|------|------------------|-------------|-------|
| `openshell.proto` | 2600 | 2206 | ~400 behind |
| `inference.proto` | exists | missing | entirely absent |
| `options.proto` | 35 lines | 14 lines | missing AuthorizationRule |
| `datamodel.proto` | needs check | needs check | likely behind |
| `sandbox.proto` | needs check | needs check | likely behind |

## Approaches Considered

### A: Full proto copy + buf regenerate
- Copy all proto files from `NVIDIA/OpenShell/proto/` to `proto/`
- Add `inference.proto` and wire it into `buf.gen.yaml`
- Run `buf generate` to regenerate all stubs
- Fix any compilation issues from new fields/types
- Pros: clean, guaranteed in sync, no drift
- Cons: may break existing converters if field types changed

### B: Selective diff + targeted update
- Diff each proto file, apply only the changes needed for workspace scoping + inference
- Skip auth annotation changes (they're gateway-internal metadata, SDK doesn't need them for client code)
- Pros: smaller diff, less breakage
- Cons: drift accumulates, auth annotations are harmless to include

## Decision

**Approach A: Full proto copy.** The auth annotations are harmless (unused by client code but nice for documentation). A clean sync prevents future drift and makes subsequent feature work straightforward. The SDK already uses buf, so regeneration is a single `make proto` command.

## Key Requirements

1. Copy all 5 proto files from upstream: `openshell.proto`, `inference.proto`, `options.proto`, `datamodel.proto`, `sandbox.proto`
2. Add `inference.proto` to `buf.gen.yaml` if needed
3. Run `buf generate` and fix compilation
4. Verify existing tests still pass (converter tests may need updates for new fields)
5. Do NOT add new client code in this step, only proto + stubs

## Open Questions

- Does `buf.gen.yaml` need changes to handle the new `inference.proto` package, or does buf discover it automatically from the proto directory?
- Are there any upstream proto files beyond the 5 listed that we should pick up? (`compute_driver.proto`, `gateway_interceptor.proto`, `supervisor_middleware.proto` exist upstream but are internal/operator-only)
- Check if `datamodel.proto` gained workspace-related fields in PR #2445
