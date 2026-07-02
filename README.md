# lazycommit

AI-powered Git commit message generator. It reads your staged diff, asks an
LLM through any OpenAI-compatible API, and prints clean commit message
suggestions — one per line, ready to pipe into lazygit, fzf, or any TUI menu.

> [!IMPORTANT]
> **v2 is a full rewrite and a breaking change.** The v1 configuration format
> and providers (opencode, Claude Code CLI, Copilot, Gemini) are gone. v2
> talks to exactly one backend family: any endpoint speaking the OpenAI
> chat-completions protocol. Run `lazycommit config set` to start fresh.

## Features

- Suggests a configurable number of commit messages from `git diff --cached`
- Suggests pull request titles from the merge-base diff against a target branch
- Works with any OpenAI-compatible endpoint: OpenAI, Ollama (local, keyless), OpenRouter, LM Studio, enterprise proxies
- Model fallback chain, request retry, and timeouts built in
- Any output language (English, Arabic, Korean, ...)
- Plain-line output designed for piping into TUI menus

## Installation

```bash
go install github.com/m7medvision/lazycommit/v2@latest
```

Or build from source:

```bash
git clone https://github.com/m7medvision/lazycommit.git
cd lazycommit
make build
```

## Quick start

```bash
lazycommit config set   # choose model, endpoint, key, language
git add .
lazycommit commit
```

## CLI

- `lazycommit commit` — prints commit message suggestions for the staged diff, one per line.
- `lazycommit pr <target-branch>` — prints pull request title suggestions for the diff against `<target-branch>`.
- `lazycommit config set` — interactive setup (model, endpoint, API key, language).
- `lazycommit config get` — shows the active backend, model, and language; API keys are masked.

Exit behavior:

- No staged changes: prints `No staged changes to commit.` and exits 0.
- Configuration or backend errors: message on stderr, non-zero exit, stdout stays clean.

## Configuration

Two files, deliberately split:

### 1. Backend settings — `~/.config/lazycommit/config.yaml`

API keys and endpoints. **Do not commit this file.** It is written with
owner-only permissions.

```yaml
active_backend: openai-compatible
backends:
  openai-compatible:
    model: gpt-4o-mini
    api_key: "$OPENAI_API_KEY" # plain value or $ENV_VAR reference
    # base_url: https://api.openai.com/v1   # optional, default is official OpenAI
    # fallback_models:                      # tried in order when the model fails
    #   - gpt-4o
```

### 2. Prompt settings — `~/.config/lazycommit/prompts.yaml`

Shareable, safe for dotfiles:

```yaml
language: English
num_suggestions: 10
# system_message: ...
# commit_message_template: "... %s"   # %s is replaced by the diff
# pr_title_template: "... %s"
```

Any repository can override prompt settings with a `lazycommit.prompts.yaml`
in its root; unset fields fall through to the global file, then to built-in
defaults:

```yaml
# my-korean-project/lazycommit.prompts.yaml
language: Korean
num_suggestions: 5
```

### Endpoint examples

**Ollama (local, no key):**

```yaml
active_backend: openai-compatible
backends:
  openai-compatible:
    model: llama3.1:8b
    base_url: http://localhost:11434/v1
```

**OpenRouter:**

```yaml
active_backend: openai-compatible
backends:
  openai-compatible:
    model: openai/gpt-4o-mini
    api_key: "$OPENROUTER_API_KEY"
    base_url: https://openrouter.ai/api/v1
```

## Integration with TUI Git clients

`lazycommit commit` prints plain lines, so it plugs directly into menu UIs.

### fzf

```bash
git add .
lazycommit commit | fzf --prompt='Pick commit> ' | xargs -r -I {} git commit -m "{}"
```

### Lazygit

Add to `~/.config/lazygit/config.yml`:

```yaml
customCommands:
  - key: "<c-a>"
    description: "pick AI commit"
    command: 'git commit -m "{{.Form.Msg}}"'
    context: "files"
    prompts:
      - type: "menuFromCommand"
        title: "AI commits"
        key: "Msg"
        command: "lazycommit commit"
        filter: "^(?P<raw>.+)$"
        valueFormat: "{{ .raw }}"
        labelFormat: "{{ .raw | green }}"
```

Variant that lets you edit the message before committing:

```yaml
  - key: "<c-b>"
    description: "Pick AI commit (edit before committing)"
    context: "files"
    command: >
      bash -c 'msg="{{.Form.Msg}}"; echo "$msg" > .git/COMMIT_EDITMSG && ${EDITOR:-nvim} .git/COMMIT_EDITMSG && if [ -s .git/COMMIT_EDITMSG ]; then
        git commit -F .git/COMMIT_EDITMSG;
      else
        echo "Commit message is empty, commit aborted.";
      fi'
    prompts:
      - type: "menuFromCommand"
        title: "AI commits"
        key: "Msg"
        command: "lazycommit commit"
        filter: "^(?P<raw>.+)$"
        valueFormat: "{{ .raw }}"
        labelFormat: "{{ .raw | green }}"
    output: terminal
```

## Troubleshooting

- `No staged changes to commit.` — run `git add` first.
- `has no model configured` — run `lazycommit config set`.
- `environment variable X is not set` — your config references `$X`; export it or store the key directly.
- Found v1 config note — v2 uses a new format; run `lazycommit config set` once and delete the old `~/.config/.lazycommit.yaml`.

## License

MIT
