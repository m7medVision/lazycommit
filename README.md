# lazycommit

AI-powered Git commit message generator that analyzes your staged changes and outputs conventional commit messages.

## Features

- Generates 10 commit message suggestions from your staged diff
- Providers: GitHub Copilot (default), OpenAI
- Interactive config to pick provider/model and set keys
- Simple output suitable for piping into TUI menus (one message per line)

## Installation

```bash
go install github.com/m7medvision/lazycommit@latest
```

Or build from source:

```bash
git clone https://github.com/m7medvision/lazycommit.git
cd lazycommit
go build -o lazycommit main.go
```

## CLI

- Root command: `lazycommit`
- Subcommands:
  - `lazycommit commit` — prints 10 suggested commit messages to stdout, one per line, based on `git diff --cached`.
  - `lazycommit config get` — prints the active provider and model.
  - `lazycommit config set` — interactive setup for provider, API key, and model.

Exit behaviors:
- If no staged changes: prints "No staged changes to commit." and exits 0.
- On config/LLM errors: prints to stderr and exits non‑zero.

### Examples

Generate suggestions after staging changes:

```bash
git add .
lazycommit commit
```

Pipe the first suggestion to commit (bash example):

```bash
MSG=$(lazycommit commit | sed -n '1p')
[ -n "$MSG" ] && git commit -m "$MSG"
```

Pick interactively with `fzf`:

```bash
git add .
lazycommit commit | fzf --prompt='Pick commit> ' | xargs -r -I {} git commit -m "{}"
```

## Configuration

- Config directory (`~/.config/.lazycommit.yaml`

```yaml
active_provider: copilot # default if a GitHub token is found
providers:
  copilot:
    api_key: "$GITHUB_TOKEN"   # Uses GitHub token; token is exchanged internally
    model: "gpt-4o"            # or "openai/gpt-4o"; both accepted
    # endpoint_url: "https://api.githubcopilot.com"  # Optional - uses default if not specified
  openai:
    api_key: "$OPENAI_API_KEY"
    model: "gpt-4o"
    # endpoint_url: "https://api.openai.com/v1"  # Optional - uses default if not specified
  # Custom provider example (e.g., local Ollama):
  # local:
  #   api_key: "not-needed"
  #   model: "llama3.1:8b"
  #   endpoint_url: "http://localhost:11434/v1"
```

### Custom Endpoints

You can configure custom API endpoints for any provider, which is useful for:
- **Local AI models**: Ollama, LM Studio, or other local inference servers
- **Enterprise proxies**: Internal API gateways or proxy servers
- **Alternative providers**: Any OpenAI-compatible API endpoint

The `endpoint_url` field is optional. If not specified, the official endpoint for that provider will be used.

#### Examples

**Ollama (local):**
```yaml
active_provider: openai  # Use openai provider for Ollama compatibility
providers:
  openai:
    api_key: "ollama"  # Ollama doesn't require real API keys
    model: "llama3.1:8b"
    endpoint_url: "http://localhost:11434/v1"
```

<!-- **Z.AI (GLM models):** -->
<!-- ```yaml -->
<!-- active_provider: openai -->
<!-- providers: -->
<!--   openai: -->
<!--     api_key: "$ZAI_API_KEY" -->
<!--     model: "glm-4.6" -->
<!--     endpoint_url: "https://api.z.ai/api/paas/v4/" -->
<!-- ``` -->

## Integration with TUI Git clients

Because `lazycommit commit` prints plain lines, it plugs nicely into menu UIs.

### Lazygit custom command

Add this to `~/.config/lazygit/config.yml`:

```yaml
customCommands:
  - key: "<c-a>" # ctrl + a
    description: "pick AI commit"
    command: 'git commit -m "{{.Form.Msg}}"'
    context: "files"
    prompts:
      - type: "menuFromCommand"
        title: "ai Commits"
        key: "Msg"
        command: "lazycommit commit"
        filter: '^(?P<raw>.+)$'
        valueFormat: "{{ .raw }}"
        labelFormat: "{{ .raw | green }}"
```

## Troubleshooting

- "No staged changes to commit." — run `git add` first.
- "API key not set" — set the appropriate key in `.lazycommit.yaml` or env var and rerun.
- Copilot errors about token exchange — ensure your GitHub token has models scope or is valid; try setting `GITHUB_TOKEN`.

## License

MIT
