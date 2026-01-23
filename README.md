# lazycommit

AI-powered Git commit message generator that analyzes your staged changes and outputs conventional commit messages.

<video src="https://github-production-user-asset-6210df.s3.amazonaws.com/88824957/518189972-f9819b7b-f33b-4544-9d65-ffee2b7c4244.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAVCODYLSA53PQK4ZA%2F20251124%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20251124T154151Z&X-Amz-Expires=300&X-Amz-Signature=9ad6523cf92ecbe4fad3b218333c036f3a9a56c88bb630a4d53239c4b20ffa78&X-Amz-SignedHeaders=host" controls title="demo">
    Your browser does not support the video tag.
</video>


## Features

- Generates configurable number of commit message suggestions from your staged diff
-  Generates 10 pull request titles based on the diff between the current branch and a target branch
- Providers: GitHub Copilot (default), OpenAI, Anthropic (Claude Code CLI)
- Multi-language support: Any language (English, Arabic, Korean, etc.)
- Interactive config to pick provider/model/language and set keys
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
  - `lazycommit pr <target-branch>` — prints 10 suggested pull request titles to stdout, one per line, based on diff between current branch and `<target-branch>`.
  - `lazycommit config get` — prints the active provider, model and language.
  - `lazycommit config set` — interactive setup for provider, API key,  model, and language.

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

Generate PR titles against `main` branch:

```bash
lazycommit pr main
```

## Configuration

lazycommit uses a two-file configuration system to separate sensitive provider settings from shareable prompt configurations:

### 1. Provider Configuration (`~/.config/.lazycommit.yaml`)
Contains API keys, tokens, and provider-specific settings. **Do not share this file.**

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
  anthropic:
    model: "claude-haiku-4-5"  # Uses Claude Code CLI - no API key needed
    num_suggestions: 10        # Number of commit suggestions to generate
```

> [!NOTE]
> `.lazycommit.yaml: language` is removed and please use `.lazycommit.prompts.yaml` instead.

### 2. Prompt Configuration (`~/.config/.lazycommit.prompts.yaml`)
Contains prompt templates and message configurations. **Safe to share in dotfiles and Git.**

```yaml
language: English # commit message language (e.g., "English", "Arabic", "Korean")
system_message: "You are a helpful assistant that generates git commit messages, and pull request titles."
commit_message_template: "Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s"
pr_title_template: "Based on the following git diff, generate 10 pull request title suggestions. Each title should be on a new line, without any numbering or bullet points:\n\n%s"
```

### Per-Repository Configuration

You can override the prompt configuration on a per-repository basis by creating a `.lazycommit.prompts.yaml` file in the root of your git repository. This is useful for projects that require different languages or commit message formats.

If a field is missing in your repository-local configuration, the value from the global configuration will be used.

Example `.lazycommit.prompts.yaml` for a Korean project:
```yaml
language: Korean
commit_message_template: "Based on the following git diff, generate 5 conventional commit messages:\n\n%s"
```
This file is automatically created on first run in the global config directory with sensible defaults:

```yaml
system_message: "You are a helpful assistant that generates git commit messages, and pull request titles."
commit_message_template: "Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s"
pr_title_template: "Based on the following git diff, generate 10 pull request title suggestions. Each title should be on a new line, without any numbering or bullet points:\n\n%s"
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

### Language Configuration

lazycommit supports generating commit messages in any language. Set the `language` field in your prompt config (`.lazycommit.prompts.yaml`):

```yaml
language: Spanish
# or
language: Arabic
# or
language: English  # (default)
```

You can also configure it interactively:

```bash
lazycommit config set  # Select language in the interactive menu
```

The language setting automatically instructs the AI to generate commit messages in the specified language, regardless of the provider used.

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

This config will allows you to edit the commit message after picking from lazycommit suggestions.
```yaml
  - key: "<c-b>" # ctrl + b
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
        title: "ai Commits"
        key: "Msg"
        command: "lazycommit commit"
        filter: '^(?P<raw>.+)$'
        valueFormat: "{{ .raw }}"
        labelFormat: "{{ .raw | green }}"
    output: terminal
```



### Commitizen

First, install the Commitizen plugin:

```bash
pip install cz-lazycommit
# or if you are using Arch Linux:
uv tool install commitizen --with cz-lazycommit
```

Then use the plugin with the following command:

```bash
git cz --name cz_lazycommit commit
```

If you are using Commitizen with Lazygit, you can add this custom command:

```yaml
  - key: "C"
    command: "git cz --name cz_lazycommit commit"
    description: "Commit with Commitizen"
    context: "files"
    loadingText: "Opening Commitizen commit tool"
    output: terminal
```


## Troubleshooting

- "No staged changes to commit." — run `git add` first.
- "API key not set" — set the appropriate key in `.lazycommit.yaml` or env var and rerun.
- Copilot errors about token exchange — ensure your GitHub token has models scope or is valid; try setting `GITHUB_TOKEN`.

## License

MIT
