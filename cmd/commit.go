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

// CommitProvider defines the interface for generating commit messages
type CommitProvider interface {
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)
	GenerateCommitMessages(ctx context.Context, diff string) ([]string, error)
}

func init() {
	RootCmd.AddCommand(commitCmd)
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate commit message suggestions",
	Long:  `Analyzes your staged changes and generates a list of 10 conventional commit message suggestions.`,
	Run: func(cmd *cobra.Command, args []string) {
		diff, err := git.GetStagedDiff()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting staged diff: %v\n", err)
			os.Exit(1)
		}

		if diff == "" {
			fmt.Println("No staged changes to commit.")
			return
		}

		var aiProvider CommitProvider

		providerName := config.GetProvider()
		apiKey, err := config.GetAPIKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting API key: %v\n", err)
			os.Exit(1)
		}

		var model string
		if providerName == "copilot" || providerName == "openai" {
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
		default:
			// Default to copilot if provider is not set or unknown
			aiProvider = provider.NewCopilotProvider(apiKey, endpoint)
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

		for _, msg := range commitMessages {
			fmt.Println(msg)
		}
	},
}
