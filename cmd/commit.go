package cmd

import (
	"github.com/spf13/cobra"
)

func newCommitCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Suggest commit messages for the staged diff, one per line",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			printV1Hint(cmd, deps)

			uc, err := deps.NewCommitUC()
			if err != nil {
				return err
			}
			res, err := uc.Execute(cmd.Context())
			if err != nil {
				return err
			}
			if res.NoChanges {
				cmd.Println("No staged changes to commit.")
				return nil
			}
			for _, s := range res.Suggestions {
				cmd.Println(s.String())
			}
			return nil
		},
	}
}
