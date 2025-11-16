package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/rm-hull/git-commit-summary/internal/app"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/setup"
	"github.com/rm-hull/git-commit-summary/internal/ui"
	"github.com/spf13/cobra"
)

func main() {
	cfg, err := config.Load()
	handleError(err)

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration is not valid, running setup wizard...\n%v\n", err)
		newCfg, err := setup.Run(cfg)
		if err != nil {
			if errors.Is(err, huh.ErrAborted) {
				handleError(interfaces.ErrAborted)
			} else {
				handleError(errors.Wrap(err, "failed to run setup wizard"))
			}
		}
		if err := newCfg.Save(); err != nil {
			handleError(errors.Wrap(err, "failed to save new configuration"))
		}
		fmt.Println("Configuration saved successfully.")
		cfg = newCfg
	}

	var userMessage string
	var llmProvider string

	rootCmd := &cobra.Command{
		Use:   "git-commit-summary",
		Short: "Generate a commit summary using Gemini or OpenAI",
		Run: func(cmd *cobra.Command, args []string) {
			version, _ := cmd.Flags().GetBool("version")
			if version {
				fmt.Println(versioninfo.Short())
				os.Exit(0)
			}

			if cmd.Flags().Changed("llm-provider") {
				cfg.LLMProvider = llmProvider
			}

			ctx := context.Background()

			provider, err := llmprovider.NewProvider(ctx, cfg)
			handleError(err)

			application := app.NewApp(provider, git.NewClient(), cfg.Prompt)
			err = application.Run(ctx, userMessage)
			if err != nil {
				handleError(err)
			}
		},
	}

	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	rootCmd.PersistentFlags().StringVarP(&userMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().StringVarP(&llmProvider, "llm-provider", "", cfg.LLMProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")

	_ = rootCmd.Execute()
}

func handleError(err error) {
	if err != nil {
		if errors.Is(err, interfaces.ErrAborted) {
			fmt.Println(ui.BoldRed.Render("ABORTED!"))
			os.Exit(0)
		} else {
			prefix := ui.BoldRed.Render("ERROR:")
			fmt.Fprintf(os.Stderr, "%s %v\n", prefix, err)
			os.Exit(1)
		}
	}
}
