# Quickstart: Upstream PR Preparation

**Date**: 2026-07-11 | **Feature**: 019-upstream-pr

## Prerequisites

- Go 1.25+ installed
- `protoc` 29.6+ and Go protoc plugins (managed via mise)
- `gh` CLI authenticated with GitHub
- `rhuss/OpenShell` fork synced with `NVIDIA/OpenShell` main
- Access to create repos under `github.com/rhuss/`

## Execution Order

### Step 1: Create PR branch in fork

```bash
cd /path/to/rhuss/OpenShell
git fetch upstream
git checkout -b go-sdk upstream/main
```

### Step 2: Copy SDK source with module path rewrite

```bash
# Copy source directories
cp -r /path/to/openshell-sdk-go/openshell sdk/go/openshell
cp -r /path/to/openshell-sdk-go/proto sdk/go/proto
cp /path/to/openshell-sdk-go/{go.mod,go.sum,Makefile,mise.toml} sdk/go/
cp -r /path/to/openshell-sdk-go/specs sdk/go/specs

# Rewrite module path
OLD="github.com/rhuss/openshell-sdk-go"
NEW="github.com/NVIDIA/OpenShell/sdk/go"
find sdk/go -name '*.go' -exec sed -i.bak "s|$OLD|$NEW|g" {} +
sed -i.bak "s|$OLD|$NEW|g" sdk/go/go.mod sdk/go/mise.toml
find sdk/go -name '*.bak' -delete

# Regenerate proto bindings with new module path, then tidy
cd sdk/go && mise run proto:gen && go mod tidy && cd ../..
```

### Step 3: Verify build and tests

```bash
cd sdk/go
go build ./...
go test ./...
```

### Step 4: Extract examples

```bash
gh repo create rhuss/openshell-examples --public
# Clone, add examples, push (details in tasks)
```

### Step 5: Create Fern docs

Create 4 MDX files under `docs/sdks/go/` and add navigation entry to
`docs/index.yml`.

### Step 6: Add CI job

Add Go job to `.github/workflows/branch-checks.yml` and create
`tasks/go.toml` with proto generation task.

### Step 7: Squash and open PR

```bash
git add -A
git commit -s -m "feat(sdk): add Go SDK for OpenShell" # Full message per research.md R7
gh pr create --draft --base main --repo NVIDIA/OpenShell \
  --title "feat(sdk): add Go SDK for OpenShell" \
  --body-file pr-description.md
```

## Verification Checklist

- [ ] `go build ./...` passes in `sdk/go/`
- [ ] `go test ./...` passes in `sdk/go/`
- [ ] Zero references to `github.com/rhuss/openshell-sdk-go` in `sdk/go/`
- [ ] `rhuss/openshell-examples` compiles against upstream module
- [ ] `docs/index.yml` includes SDK section
- [ ] PR excludes brainstorm/, .specify/, .claude/, CLAUDE.md, AGENTS.md
- [ ] PR references issue #2044
- [ ] Commit has `Signed-off-by` trailer
