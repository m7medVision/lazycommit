# lazycommit

AI-powered Git commit message generator that analyzes your staged changes and generates conventional commit messages.

## Features

- Generates commit messages using AI (supports OpenAI, OpenRouter, and GitHub Copilot)
- Interactive prompts for accepting, editing, or canceling commit messages
- Easy to extend with other AI providers
- Built with Go for performance and portability

## Installation

```bash
go install github.com/m7medvision/lazycommit@latest
```

Or clone the repository and build from source:

```bash
git clone https://github.com/m7medvision/lazycommit.git
cd lazycommit
go build -o lazycommit main.go
```

## Usage

Stage your changes first:

```bash
git add .
```

Then generate a commit message:

```bash
lazycommit commit
```

You can also run the interactive configuration to set your preferred provider and model:

```bash
lazycommit config set
```

For automatic commit without prompting:

```bash
lazycommit commit -a
```

## Configuration

- Config directory (`~/.config/.lazycommit.yaml`)

```yaml
active_provider: openai
providers:
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-4o"
  openrouter:
    api_key: "$env-your-openrouter-api-key"
    model: "anthropic/claude-3-opus"
  copilot:
    # No API key needed, it uses the GitHub CLI token
    model: "gpt-4o"
```

## Development

### Project Structure

```
lazycommit/
├── cmd/
│   └── commit.go
│   └── root.go
├── internal/
│   ├── git/
│   │   └── git.go
│   ├── provider/
│   │   ├── provider.go
│   │   ├── copilot.go
│   │   └── models/
│   │       └── models.go
│   └── config/
│       └── config.go
├── main.go
└── go.mod
```

### Building

```bash
go build -o lazycommit main.go
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [go-git](https://github.com/go-git/go-git) - Git implementation in Go
- [Survey](https://github.com/AlecAivazis/survey) - Interactive prompts

## License

MIT
