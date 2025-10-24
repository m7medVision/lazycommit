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

**LM Studio (local):**
```yaml
active_provider: openai
providers:
  openai:
    api_key: "not-needed"
    model: "local-model"
    endpoint_url: "http://localhost:1234/v1"
```

**Enterprise Proxy:**
```yaml
active_provider: openai
providers:
  openai:
    api_key: "$ENTERPRISE_OPENAI_KEY"
    model: "gpt-4o"
    endpoint_url: "https://ai-proxy.company.com/openai/v1"
```

**Z.AI (GLM models):**
```yaml
active_provider: openai
providers:
  openai:
    api_key: "$ZAI_API_KEY"
    model: "glm-4.6"
    endpoint_url: "https://api.z.ai/api/paas/v4/"
```

**Available GLM models:**
- `glm-4.6` - Latest GLM model with strong reasoning capabilities
- `glm-4.5` - High-performance model with thinking mode support
- `glm-4.5-air` - Lightweight version of GLM-4.5

To get started with Z.AI:
1. Register at [Z.AI Open Platform](https://z.ai/model-api)
2. Create an API key in the [API Keys](https://z.ai/manage-apikey/apikey-list) management page
3. Set the `ZAI_API_KEY` environment variable or add it to your config
4. Configure lazycommit as shown above

Notes:
- Copilot: requires a GitHub token with models scope. The tool can also discover IDE Copilot tokens, but models scope is recommended.
- Environment variable references are supported by prefixing with `$` (e.g., `$OPENAI_API_KEY`).
- Custom endpoints must be OpenAI-compatible for proper functionality.
- Endpoint URLs are validated to ensure they use HTTP/HTTPS protocols and have valid hosts.

### Configure via CLI

```bash
lazycommit config set     # interactive provider/model/key/endpoint picker
lazycommit config get     # show current provider/model/endpoint
```

The interactive setup will now prompt for custom endpoint URLs as well. You can leave this field empty to use the default endpoint for the selected provider.

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

Tips:
- For `lazycommit commit`, you can omit `filter` and just use `valueFormat: "{{ .raw }}"` and `labelFormat: "{{ .raw | green }}"`.
- If you pipe a numbered list tool (e.g., `bunx bunnai`), keep the regex groups `number` and `message` as shown.

## Providers and models

- Copilot (default when a GitHub token is available): uses `gpt-4o` unless overridden. Accepts `openai/gpt-4o` and normalizes it to `gpt-4o`.
- OpenAI: choose from models defined in the interactive picker (e.g., gpt‑4o, gpt‑4.1, o3, o1, etc.).

## How it works

- Reads `git diff --cached`.
- Sends a single prompt to the selected provider to generate 10 lines.
- Prints the lines exactly, suitable for piping/selecting.

## Troubleshooting

- "No staged changes to commit." — run `git add` first.
- "API key not set" — set the appropriate key in `.lazycommit.yaml` or env var and rerun.
- Copilot errors about token exchange — ensure your GitHub token has models scope or is valid; try setting `GITHUB_TOKEN`.

## License

MIT
