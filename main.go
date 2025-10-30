package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/gookit/color"
	"github.com/joho/godotenv"
	"github.com/rm-hull/git-commit-summary/internal/app" // New import
	"github.com/spf13/cobra"
)

//go:embed prompt.md
var prompt string

func main() {
	configFile, err := xdg.ConfigFile("git-commit-summary/config.env")
	handleError(err)

	_ = godotenv.Load(configFile)
	_ = godotenv.Overload(".env")

	defaultProvider := os.Getenv("LLM_PROVIDER")
	if defaultProvider == "" {
		defaultProvider = "google"
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

			ctx := context.Background()

			application, err := app.NewApp(ctx, llmProvider, prompt)
			if err != nil {
				handleError(err)
			}
			handleError(application.Run(ctx, userMessage))
		},
	}

	rootCmd.PersistentFlags().StringVarP(&userMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	rootCmd.PersistentFlags().StringVarP(&llmProvider, "llm-provider", "", defaultProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")

	_ = rootCmd.Execute()
}

func handleError(err error) {
	if err != nil {
		color.Fprintf(os.Stderr, "<fg=red;op=bold>ERROR:</> %v\n", err)
		os.Exit(1)
	}
}
