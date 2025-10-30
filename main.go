package main

import (
	"context"
	"fmt"
	"os"

	"github.com/earthboundkid/versioninfo/v2"
	"github.com/gookit/color"
	"github.com/rm-hull/git-commit-summary/internal/app"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	cfg, err := config.Load()
	handleError(err)

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

			application, err := app.NewApp(ctx, cfg)
			if err != nil {
				handleError(err)
			}
			handleError(application.Run(ctx, userMessage))
		},
	}

	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	rootCmd.PersistentFlags().StringVarP(&userMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().StringVarP(&llmProvider, "llm-provider", "", cfg.LLMProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")

	_ = rootCmd.Execute()
}

func handleError(err error) {
	if err != nil {
		color.Fprintf(os.Stderr, "<fg=red;op=bold>ERROR:</> %v\n", err)
		os.Exit(1)
	}
}
