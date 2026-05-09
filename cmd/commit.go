package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/m7medvision/lazycommit/internal/git"
	"github.com/m7medvision/lazycommit/internal/provider"
	"github.com/spf13/cobra"
)

// CommitProvider defines the interface for generating commit messages
type CommitProvider interface {
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)
	GenerateCommitMessages(ctx context.Context, diff string) ([]string, error)
}

func init() {
	RootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVarP(&commitProviderFlag, "provider", "p", "", "Provider override: opencode, openai, copilot, anthropic, gemini")
	commitCmd.Flags().StringVarP(&commitModelFlag, "model", "m", "", "Model override for selected provider")
	commitCmd.Flags().IntVarP(&commitGenerateFlag, "generate", "g", 0, "Number of commit message suggestions to generate")
	commitCmd.Flags().StringVarP(&commitLanguageFlag, "lang", "l", "", "Language override for generated commit messages")
	commitCmd.Flags().BoolVarP(&commitEmojiFlag, "emoji", "e", false, "Prefix generated commit messages with gitmoji")
	commitCmd.Flags().BoolVarP(&commitMessageOnlyFlag, "message-only", "o", false, "Print only the first generated message")
	commitCmd.Flags().BoolVar(&commitStageAllFlag, "stage-all", false, "Stage all tracked, deleted, and untracked changes before generating commit messages")
	commitCmd.Flags().BoolVarP(&commitSilentEmptyFlag, "silent-empty", "n", false, "Stay silent when there are no staged changes")
	commitCmd.Flags().BoolVarP(&commitDebugFlag, "debug", "d", false, "Show debug diagnostics")
}

var (
	commitProviderFlag    string
	commitModelFlag       string
	commitGenerateFlag    int
	commitLanguageFlag    string
	commitEmojiFlag       bool
	commitMessageOnlyFlag bool
	commitStageAllFlag    bool
	commitSilentEmptyFlag bool
	commitDebugFlag       bool
)

var conventionalTypePattern = regexp.MustCompile(`^([a-z]+)(\([^)]+\))?:\s+`)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate commit message suggestions",
	Long:  `Analyzes your staged changes and generates conventional commit message suggestions.`,
	Example: `  lazycommit commit
  lazycommit commit --stage-all
  lazycommit commit -p opencode -m opencode/minimax-m2.5-free
  lazycommit commit -g 3 -l Spanish
  lazycommit commit -o`,
	Run: func(cmd *cobra.Command, args []string) {
		if commitStageAllFlag {
			hasChanges, err := git.HasChanges()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking git status: %v\n", err)
				os.Exit(1)
			}
			if !hasChanges {
				if !commitSilentEmptyFlag {
					fmt.Println("No changes to stage.")
				}
				return
			}
			if err := git.StageAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error staging changes: %v\n", err)
				os.Exit(1)
			}
			if commitDebugFlag {
				fmt.Fprintln(os.Stderr, "debug: staged all changes with git add --all")
			}
		}

		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting staged diff: %v\n", err)
			os.Exit(1)
		}

		if diff == "" {
			if !commitSilentEmptyFlag {
				fmt.Println("No staged changes to commit.")
			}
			if commitDebugFlag {
				fmt.Fprintln(os.Stderr, "debug: staged diff is empty")
			}
			return
		}

		providerName := strings.TrimSpace(config.GetProvider())
		if strings.TrimSpace(commitProviderFlag) != "" {
			providerName = strings.TrimSpace(commitProviderFlag)
		}

		if providerName == "" {
			fmt.Fprintln(os.Stderr, "Provider is empty. Set one with 'lazycommit config set' or use --provider.")
			os.Exit(1)
		}
		if !isSupportedCommitProvider(providerName) {
			fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerName)
			os.Exit(1)
		}

		var aiProvider CommitProvider

		generateCount := commitGenerateFlag
		if generateCount <= 0 {
			generateCount = config.GetNumSuggestionsForProvider(providerName)
		}
		if generateCount <= 0 {
			generateCount = 10
		}

		// API keys are not needed for CLI-backed providers.
		var apiKey string
		if providerName != "anthropic" && providerName != "gemini" && providerName != "opencode" {
			var err error
			apiKey, err = config.GetAPIKeyForProvider(providerName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting API key: %v\n", err)
				os.Exit(1)
			}
		}

		var model string
		if strings.TrimSpace(commitModelFlag) != "" {
			model = strings.TrimSpace(commitModelFlag)
		} else if providerName == "copilot" || providerName == "openai" || providerName == "anthropic" || providerName == "gemini" || providerName == "opencode" {
			var err error
			model, err = config.GetModelForProvider(providerName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting model: %v\n", err)
				os.Exit(1)
			}
		}

		endpoint, err := config.GetEndpointForProvider(providerName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting endpoint: %v\n", err)
			os.Exit(1)
		}

		if commitDebugFlag {
			fmt.Fprintf(os.Stderr, "debug: provider=%s model=%s generate=%d lang=%q emoji=%t message_only=%t silent_empty=%t\n",
				providerName, model, generateCount, commitLanguageFlag, commitEmojiFlag, commitMessageOnlyFlag, commitSilentEmptyFlag)
			fmt.Fprintf(os.Stderr, "debug: stage_all=%t\n", commitStageAllFlag)
			fmt.Fprintf(os.Stderr, "debug: diff_bytes=%d endpoint=%q\n", len(diff), endpoint)
		}

		provider.SetRuntimeCommitPromptOptions(provider.CommitPromptOptions{
			Generate: generateCount,
			Language: strings.TrimSpace(commitLanguageFlag),
			Emoji:    commitEmojiFlag,
		})
		defer provider.ResetRuntimeCommitPromptOptions()

		switch providerName {
		case "copilot":
			aiProvider = provider.NewCopilotProviderWithModel(apiKey, model, endpoint)
		case "openai":
			aiProvider = provider.NewOpenAIProvider(apiKey, model, endpoint)
		case "anthropic":
			aiProvider = provider.NewAnthropicProvider(model, generateCount)
		case "gemini":
			aiProvider = provider.NewGeminiProvider(model, generateCount)
		case "opencode":
			aiProvider = provider.NewOpencodeProvider(model, config.GetFallbackModelsForProvider(providerName), generateCount)
		}

		commitMessages, err := aiProvider.GenerateCommitMessages(context.Background(), diff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating commit messages: %v\n", err)
			os.Exit(1)
		}

		if len(commitMessages) == 0 {
			fmt.Println("No commit messages generated.")
			return
		}

		if generateCount > 0 && len(commitMessages) > generateCount {
			commitMessages = commitMessages[:generateCount]
		}

		commitMessages = applyOutputOverrides(commitMessages, commitEmojiFlag)

		if commitDebugFlag {
			fmt.Fprintf(os.Stderr, "debug: generated_messages=%d\n", len(commitMessages))
		}

		if commitMessageOnlyFlag {
			fmt.Println(commitMessages[0])
			return
		}

		for _, msg := range commitMessages {
			fmt.Println(msg)
		}
	},
}

func isSupportedCommitProvider(providerName string) bool {
	switch providerName {
	case "copilot", "openai", "anthropic", "gemini", "opencode":
		return true
	default:
		return false
	}
}

func applyOutputOverrides(messages []string, addEmoji bool) []string {
	out := make([]string, 0, len(messages))
	for _, msg := range messages {
		updated := msg
		if addEmoji {
			updated = ensureGitmojiPrefix(updated)
		}
		out = append(out, updated)
	}
	return out
}

func ensureGitmojiPrefix(msg string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return msg
	}
	if hasLeadingEmoji(trimmed) {
		return msg
	}
	matches := conventionalTypePattern.FindStringSubmatch(trimmed)
	if len(matches) < 2 {
		return msg
	}
	emojiByType := map[string]string{
		"feat":     "✨",
		"fix":      "🐛",
		"refactor": "♻️",
		"perf":     "⚡️",
		"docs":     "📝",
		"style":    "🎨",
		"test":     "🧪",
		"chore":    "🔧",
		"ci":       "👷",
		"build":    "📦",
		"revert":   "⏪️",
		"security": "🔒️",
	}
	if emoji, ok := emojiByType[matches[1]]; ok {
		return emoji + " " + trimmed
	}
	return msg
}

func hasLeadingEmoji(s string) bool {
	for _, r := range strings.TrimSpace(s) {
		return isEmojiRune(r)
	}
	return false
}

func isEmojiRune(r rune) bool {
	switch {
	case r >= 0x1F000 && r <= 0x1FAFF:
		return true
	case r >= 0x2600 && r <= 0x27BF:
		return true
	case r == 0x00A9 || r == 0x00AE || r == 0x3030:
		return true
	default:
		return false
	}
}
