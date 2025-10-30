package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	// "github.com/Delta456/box-cli-maker/v2"
	"github.com/adrg/xdg"
	"github.com/briandowns/spinner"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/galactixx/stringwrap"
	"github.com/gookit/color"
	"github.com/joho/godotenv"
	"github.com/rm-hull/git-commit-summary/internal"
	"github.com/rm-hull/git-commit-summary/internal/git"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/spf13/cobra"
)

func handleError(err error) {
	if err != nil {
		color.Fprintf(os.Stderr, "<fg=red;op=bold>ERROR:</> %v\n", err)
		os.Exit(1)
	}
}

//go:embed prompt.md
var prompt string
var userMessage string
var llmProvider string

func main() {
	configFile, err := xdg.ConfigFile("git-commit-summary/config.env")
	handleError(err)

	_ = godotenv.Load(configFile)
	_ = godotenv.Overload(".env")

	defaultProvider := os.Getenv("LLM_PROVIDER")
	if defaultProvider == "" {
		defaultProvider = "google"
	}

	rootCmd := &cobra.Command{
		Use:   "git-commit-summary",
		Short: "Generate a commit summary using Gemini or OpenAI",
		Run:   run,
	}
	rootCmd.PersistentFlags().StringVarP(&userMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	rootCmd.PersistentFlags().StringVarP(&llmProvider, "llm-provider", "", defaultProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")

	handleError(rootCmd.Execute())
}

func run(cmd *cobra.Command, args []string) {
	version, _ := cmd.Flags().GetBool("version")
	if version {
		fmt.Println(versioninfo.Short())
		os.Exit(0)
	}

	ctx := context.Background()
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = color.Render(" <magenta>Running git diff</>")
	s.Start()

	out, err := git.Diff()
	if err != nil {
		s.Stop()
		handleError(err)
	}

	if len(out) == 0 {
		s.Stop()
		handleError(errors.New("no changes are staged"))
	}

	provider, err := llmprovider.NewProvider(ctx, llmProvider)
	if err != nil {
		s.Stop()
		handleError(err)
	}

	s.Suffix = color.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", provider.Model())
	text := fmt.Sprintf(prompt, out)

	message, err := provider.Call(ctx, "", text)
	if err != nil {
		s.Stop()
		handleError(err)
	}

	s.Stop()

	if userMessage != "" {
		message = fmt.Sprintf("%s\n\n%s", userMessage, message)
	}

	wrapped, _, err := stringwrap.StringWrap(message, 72, 4, false)
	handleError(err)

	wrapped = strings.ReplaceAll(wrapped, "\n\n\n", "\n\n")
	edited, accepted, err := internal.TextArea(wrapped)
	handleError(err)

	if accepted {
		handleError(git.Commit(edited))
	} else {
		color.Println("<fg=red;op=bold>ABORTED!</>")
	}
}
