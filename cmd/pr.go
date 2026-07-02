package cmd

import (
	"github.com/spf13/cobra"
)

func newPRCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "pr <target-branch>",
		Short: "Suggest pull request titles against a target branch, one per line",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			printV1Hint(cmd, deps)

			uc, err := deps.NewPRUC()
			if err != nil {
				return err
			}
			res, err := uc.Execute(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if res.NoChanges {
				cmd.Printf("No changes against %s.\n", args[0])
				return nil
			}
			for _, s := range res.Suggestions {
				cmd.Println(s.String())
			}
			return nil
		},
	}
}
