---
name: openshell-proto-sync
description: >-
  Use when syncing the Go SDK (openshell-sdk-go) with upstream proto
  changes from NVIDIA/OpenShell, after upstream merges proto-related
  PRs, when SDK interfaces are out of date with proto definitions,
  or when the user asks to "sync protos", "update from upstream",
  or "gap analysis on proto".
---

# OpenShell Proto Sync

Sync proto files from NVIDIA/OpenShell into this SDK repo, regenerate
Go bindings, run gap analysis, and implement missing functionality.

## Repositories

| Repo | Path | Role |
|------|------|------|
| **This SDK** | `/Users/rhuss/Work/projects/openshell-sdk-go` | Where we work. Proto source files live in `proto/`, generated Go bindings in `proto/*/` |
| **Upstream** | `/Users/rhuss/Work/projects/openshell` | NVIDIA/OpenShell fork. Source of truth for `.proto` files (in `proto/`) |

The SDK copies `.proto` files from upstream, regenerates Go bindings
locally with `mise run proto:gen`, then implements SDK coverage for
any new or changed RPCs/fields.

## Workflow

### Phase 1: Fetch Latest Upstream Protos

```bash
UPSTREAM=/Users/rhuss/Work/projects/openshell
SDK=/Users/rhuss/Work/projects/openshell-sdk-go
```

1. Fetch upstream changes in the OpenShell fork:
   ```bash
   cd $UPSTREAM && git fetch upstream && git log --oneline upstream/main -5
   ```
2. Check what proto changes exist since the SDK's last sync:
   ```bash
   # Find the SDK's last proto sync commit
   cd $SDK && git log --oneline --grep="proto" -5
   ```
3. Copy the SDK-relevant proto files from upstream:
   ```bash
   for f in openshell.proto datamodel.proto sandbox.proto options.proto inference.proto; do
     cp "$UPSTREAM/proto/$f" "$SDK/proto/$f"
   done
   ```
   Check for **new** proto files in upstream that aren't in the SDK yet:
   ```bash
   diff <(ls $UPSTREAM/proto/*.proto | xargs -n1 basename | sort) \
        <(ls $SDK/proto/*.proto | xargs -n1 basename | sort)
   ```
   If new files appear, add them to `buf.gen.yaml`'s inputs and plugin
   mappings before regenerating.

### Phase 2: Regenerate Proto Bindings

```bash
cd $SDK && mise run proto:gen
```

Commit the proto source files and regenerated `.pb.go` files together:
```
feat(proto): sync proto files from upstream OpenShell
```

### Phase 3: Gap Analysis

Compare the regenerated proto against the SDK implementation.

**3a. New proto fields in existing messages**

```bash
rg "proto3" proto/openshellv1/openshell.pb.go | wc -l
```

Compare field counts between proto structs and the SDK's converter/type
files. Focus on:
- Request structs: new fields needing SDK method parameters
- Response structs: new fields needing converter extraction
- New enums: need type aliases in `types/`

**3b. New RPC methods**

```bash
rg "func \(c \*openShellClient\)" proto/openshellv1/openshell_grpc.pb.go
```

Cross-reference against the SDK's interface definitions in
`openshell/v1/*.go`. Each RPC should map to a method on a sub-client.

**3c. New proto files / resource types**

New `.proto` files mean entirely new resource groups needing types,
converters, interfaces, and client implementations.

**3d. Changed or removed fields**

```bash
git diff proto/ --stat
```

Look for removed fields that the SDK currently relies on.

### Phase 4: Report and Triage Findings

Present a table:

```
| Category | Item | Size | Action |
|----------|------|------|--------|
| New field | FooRequest.Bar | small | Add to foo_client.go |
| New RPC | ListWorkspaces | large | New interface + client + types |
| Removed field | SandboxSpec.X | small | Remove from converter |
```

Classify each finding:

- **Small**: new field, parameter change, enum addition.
  Implement directly in Phase 5.
- **Large**: new RPC group, new resource type, new sub-client, new proto
  file, or cross-cutting change touching 5+ files.
  Escalate via Phase 4b.

#### Phase 4b: Escalate Large Features to SpecKit

If large items are found, propose:

> "The gap analysis found N large feature(s) that would benefit from
> a spec-driven workflow. Want me to create brainstorm files?"

If the user agrees, for each large feature:

1. Determine the next brainstorm number from `brainstorm/00-overview.md`.
2. Create `brainstorm/NNN-<feature-slug>.md`:
   ```markdown
   # Brainstorm: <Feature Title>

   **Date:** YYYY-MM-DD
   **Status:** active

   ## Problem Framing
   <What upstream proto change introduced this and why the SDK needs it>

   ## Proto Surface
   <New/changed proto messages, RPCs, and fields>

   ## SDK Impact
   <Which layers need work: types, converters, interface, client, tests>

   ## Open Questions
   <Anything unclear from the proto definitions alone>
   ```
3. Update `brainstorm/00-overview.md` with the new entry.
4. At the end of the run, remind the user:
   > "Created brainstorm file(s) for the large features. Run
   > `/speckit-spex-brainstorm brainstorm/NNN-<slug>.md` for each."

### Phase 5: Implement Small Items

For each approved small item, follow the SDK's layered architecture:

1. **types/**: Add or update domain types (no proto imports)
2. **internal/converter/**: Add proto-to-SDK and SDK-to-proto converters
3. **Interface file** (e.g., `sandbox.go`): Update interface signatures
4. **Client file** (e.g., `sandbox_client.go`): Implement with gRPC + converter
5. **Tests**: Add unit tests following existing table-driven patterns
6. **types_reexport.go**: Add type aliases if new types were created

### Phase 6: Verify

```bash
cd $SDK && make ci
```

This runs lint, build, test, proto-check, and docs-check. All must pass.

### Phase 7: Commit and Push

Ask the user how to land the changes:

- **Direct to main**: For small, mechanical syncs (field additions, regen only).
  ```bash
  git add -A && git commit -m "feat(proto): sync proto and add <summary>"
  git push origin main
  ```
- **Pull request**: For changes that add new functionality or touch multiple
  SDK layers.
  ```bash
  git checkout -b proto-sync-<short-slug>
  git add -A && git commit -m "feat(proto): sync proto and add <summary>"
  git push -u origin proto-sync-<short-slug>
  gh pr create --title "feat(proto): <summary>" --body "$(cat <<'EOF'
  ## Summary
  - Synced proto files from NVIDIA/OpenShell upstream/main
  - <list of changes>

  ## Gap Analysis
  <table of findings>

  ## Testing
  - `make ci` passes (lint, build, tests, proto-check)
  EOF
  )"
  ```

Report the commit SHA or PR URL back to the user.

## SDK Architecture

```
openshell/v1/
  types/           # Domain types, no proto imports
  internal/
    converter/     # Proto <-> SDK type conversion
    grpc/          # Connection management
  client.go        # ClientInterface + Client struct
  {resource}.go    # Interface definition + type re-exports
  {resource}_client.go  # gRPC implementation
```

## Conventions

- `workspace string` is always the second parameter after `ctx`
- Converters deep-copy all slices and maps at boundaries
- Tests use bufconn for in-process gRPC, testify for assertions
- Import path: `github.com/rhuss/openshell-sdk-go/...`

## Common Pitfalls

- Proto field names use snake_case; Go uses CamelCase. Converters must
  map consistently.
- Coverage test (`coverage_test.go`) detects unconverted proto fields.
  New proto fields trigger test failures until converters are updated.
- New `.proto` files need entries in `buf.gen.yaml` (inputs + mappings).
- The `proto:check` CI step diffs generated files against source. Always
  regenerate after copying new proto source files.
