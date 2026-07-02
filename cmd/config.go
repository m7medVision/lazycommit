package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/m7medvision/lazycommit/v2/internal/config"
)

func newConfigCmd(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "config",
		Short: "Inspect or change lazycommit configuration",
	}
	root.AddCommand(newConfigGetCmd(deps), newConfigSetCmd(deps))
	return root
}

func newConfigGetCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Print the effective backend, model, and language",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true

			backends, err := deps.ConfigRepo.LoadBackendsRaw()
			if err != nil {
				return err
			}
			settings := backends.Backends[backends.Active]
			prompts, err := deps.ConfigRepo.PromptSettings()
			if err != nil {
				return err
			}

			cmd.Printf("backend:  %s\n", backends.Active)
			cmd.Printf("model:    %s\n", settings.Model)
			if len(settings.FallbackModels) > 0 {
				cmd.Printf("fallback: %s\n", strings.Join(settings.FallbackModels, ", "))
			}
			if settings.BaseURL != "" {
				cmd.Printf("base_url: %s\n", settings.BaseURL)
			}
			if settings.APIKey != "" {
				cmd.Printf("api_key:  %s\n", maskSecret(settings.APIKey))
			}
			cmd.Printf("language: %s\n", prompts.Language)
			cmd.Printf("count:    %d\n", prompts.SuggestionCount)
			return nil
		},
	}
}

func newConfigSetCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Interactively choose backend, model, and language",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			in := bufio.NewScanner(cmd.InOrStdin())

			backends, err := deps.ConfigRepo.LoadBackendsRaw()
			if err != nil {
				return err
			}

			active, err := chooseBackend(cmd, in, deps.BackendNames, backends.Active)
			if err != nil {
				return err
			}
			backends.Active = active

			settings := backends.Backends[active]
			settings.Model = ask(cmd, in, fmt.Sprintf("Model [%s]: ", orNone(settings.Model)), settings.Model)
			if active == "openai-compatible" {
				settings.BaseURL = ask(cmd, in,
					fmt.Sprintf("Base URL (empty for official OpenAI) [%s]: ", orNone(settings.BaseURL)), settings.BaseURL)
				settings.APIKey = ask(cmd, in,
					fmt.Sprintf("API key, plain or $ENV_VAR [%s]: ", maskSecret(settings.APIKey)), settings.APIKey)
			}
			if backends.Backends == nil {
				backends.Backends = map[string]config.BackendSettings{}
			}
			backends.Backends[active] = settings

			if err := deps.ConfigRepo.SaveBackends(backends); err != nil {
				return err
			}

			prompts, err := deps.ConfigRepo.LoadGlobalPrompts()
			if err != nil {
				return err
			}
			language := prompts.Language
			if language == "" {
				language = "English"
			}
			prompts.Language = ask(cmd, in, fmt.Sprintf("Language [%s]: ", language), language)
			if err := deps.ConfigRepo.SavePrompts(prompts); err != nil {
				return err
			}

			cmd.Println("Configuration saved.")
			return nil
		},
	}
}

func chooseBackend(cmd *cobra.Command, in *bufio.Scanner, names []string, current string) (string, error) {
	cmd.Println("Backends:")
	for i, name := range names {
		marker := " "
		if name == current {
			marker = "*"
		}
		cmd.Printf("  %d) %s %s\n", i+1, marker, name)
	}
	answer := ask(cmd, in, fmt.Sprintf("Choose backend [%s]: ", current), current)

	if n, err := strconv.Atoi(answer); err == nil {
		if n < 1 || n > len(names) {
			return "", fmt.Errorf("choice %d out of range 1-%d", n, len(names))
		}
		return names[n-1], nil
	}
	for _, name := range names {
		if name == answer {
			return name, nil
		}
	}
	return "", fmt.Errorf("unknown backend %q (available: %s)", answer, strings.Join(names, ", "))
}

// ask prompts and returns the trimmed reply, or fallback when the reply is
// empty or stdin is closed.
func ask(cmd *cobra.Command, in *bufio.Scanner, prompt, fallback string) string {
	cmd.Print(prompt)
	if !in.Scan() {
		cmd.Println()
		return fallback
	}
	answer := strings.TrimSpace(in.Text())
	if answer == "" {
		return fallback
	}
	return answer
}

// maskSecret hides key material: $ENV references are shown as-is (they are
// not secrets), anything else keeps only its last 4 characters.
func maskSecret(key string) string {
	if key == "" {
		return "(none)"
	}
	if strings.HasPrefix(key, "$") {
		return key
	}
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
