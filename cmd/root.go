// Package cmd is the thin CLI shell: each command parses input, calls one
// use case, and prints. Business logic lives in internal/app.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/m7medvision/lazycommit/internal/app"
	"github.com/m7medvision/lazycommit/internal/config"
)

// Deps is everything the commands need, wired by the composition root. Use
// cases are built lazily so `config set` still works when the active
// backend's configuration is currently broken.
type Deps struct {
	NewCommitUC  func() (*app.GenerateCommitSuggestions, error)
	NewPRUC      func() (*app.GeneratePRTitles, error)
	ConfigRepo   *config.Repository
	BackendNames []string
	Version      string
	// V1Hint is a warning about leftover v1 configuration, printed to
	// stderr before generation commands; empty when not applicable.
	V1Hint string
}

func NewRoot(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "lazycommit",
		Short:         "AI-powered commit message and PR title suggestions",
		Version:       deps.Version,
		SilenceErrors: true,
	}
	root.AddCommand(newCommitCmd(deps), newPRCmd(deps), newConfigCmd(deps))
	return root
}

func printV1Hint(cmd *cobra.Command, deps Deps) {
	if deps.V1Hint != "" {
		cmd.PrintErrln(deps.V1Hint)
	}
}
