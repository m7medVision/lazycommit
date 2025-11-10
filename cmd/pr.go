package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/m7medvision/lazycommit/internal/git"
	"github.com/m7medvision/lazycommit/internal/provider"
	"github.com/spf13/cobra"
)

// PrProvider defines the interface for generating pull request titles
type PrProvider interface {
	GeneratePRTitle(ctx context.Context, diff string) (string, error)
	GeneratePRTitles(ctx context.Context, diff string) ([]string, error)
}

// prCmd represents the pr command
var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Generate pull request title suggestions",
	Long: `Analyzes the diff of the current branch compared to a target branch, and generates a list of 10 suggested pull request titles.

	Arguments:
  <target-branch>    The branch to compare against (e.g., main, develop)`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing required argument: <target-branch>")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected 1 but got %d", len(args))
		}
		return nil
	},
	Example: "lazycommit pr main\n  lazycommit pr develop",
	Run: func(cmd *cobra.Command, args []string) {
		diff, err := git.GetDiffAgainstBranch(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting branch comparison diff: %v\n", err)
			os.Exit(1)
		}

		if diff == "" {
			fmt.Println("No changes compared to base branch.")
			return
		}

		var aiProvider PrProvider

		providerName := config.GetProvider()

		// API key is not needed for anthropic provider (uses CLI)
		var apiKey string
		if providerName != "anthropic" {
			var err error
			apiKey, err = config.GetAPIKey()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting API key: %v\n", err)
				os.Exit(1)
			}
		}

		var model string
		if providerName == "copilot" || providerName == "openai" || providerName == "anthropic" {
			var err error
			model, err = config.GetModel()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting model: %v\n", err)
				os.Exit(1)
			}
		}

		endpoint, err := config.GetEndpoint()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting endpoint: %v\n", err)
			os.Exit(1)
		}

		switch providerName {
		case "copilot":
			aiProvider = provider.NewCopilotProviderWithModel(apiKey, model, endpoint)
		case "openai":
			aiProvider = provider.NewOpenAIProvider(apiKey, model, endpoint)
		case "anthropic":
			// Get num_suggestions from config
			numSuggestions := config.GetNumSuggestions()
			aiProvider = provider.NewAnthropicProvider(model, numSuggestions)
		default:
			// Default to copilot if provider is not set or unknown
			aiProvider = provider.NewCopilotProvider(apiKey, endpoint)
		}

		prTitles, err := aiProvider.GeneratePRTitles(context.Background(), diff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating pull request titles %v\n", err)
			os.Exit(1)
		}

		if len(prTitles) == 0 {
			fmt.Println("No PR titles generated.")
			return
		}

		for _, title := range prTitles {
			fmt.Println(title)
		}

	},
}

func init() {
	RootCmd.AddCommand(prCmd)
}
