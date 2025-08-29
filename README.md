# lazycommit

AI-powered Git commit message generator that analyzes your staged changes and outputs conventional commit messages.

## Features

- Generates 10 commit message suggestions from your staged diff
- Providers: GitHub Copilot (default), OpenAI, OpenRouter
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
  openai:
    api_key: "$OPENAI_API_KEY"
    model: "gpt-4o"
  openrouter:
    api_key: "$OPENROUTER_API_KEY" # or a literal key
    model: "openai/gpt-4o"         # OpenRouter model IDs, e.g. anthropic/claude-3.5-sonnet
```

Notes:
- Copilot: requires a GitHub token with models scope. The tool can also discover IDE Copilot tokens, but models scope is recommended.
- Environment variable references are supported by prefixing with `$` (e.g., `$OPENAI_API_KEY`).

### Configure via CLI

```bash
lazycommit config set     # interactive provider/model/key picker
lazycommit config get     # show current provider/model
```

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
- OpenRouter: pick from OpenRouter-prefixed IDs (e.g., `openai/gpt-4o`, `anthropic/claude-3.5-sonnet`). Extra headers are set automatically.

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
