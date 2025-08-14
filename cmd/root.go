package cmd

import (
	"fmt"
	"os"

	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "lazycommit",
	Short: "lazycommit generates AI-powered git commit messages",
	Long: `lazycommit uses AI to analyze your staged changes and
generates a conventional commit message for you.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)
	RootCmd.AddCommand(configCmd)
}
