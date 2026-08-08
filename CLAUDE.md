@AGENTS.md

## Codex OpenAI Key (this project only)

This project uses a dedicated OpenAI key for Codex calls. Before invoking Codex commands (`/codex:review`, `/codex:rescue`), swap in the project key:

```bash
export OPENAI_API_KEY="$OPENSHELL_OPENAI_API_KEY"
```

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
at specs/025-reverse-port-forwarding/plan.md
<!-- SPECKIT END -->
