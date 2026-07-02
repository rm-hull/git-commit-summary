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

	var userMessage string
	var llmProvider string
	var runSetupWizard *bool
	var showVersion *bool
	var yoloMode *bool
	var addAll *bool
	var skipCI *bool
	var hint string
	var installHook *bool
	var uninstallHook *bool

	rootCmd := &cobra.Command{
		Use:   "git-commit-summary",
		Short: "Generate a commit summary using Gemini or OpenAI",
		Run: func(cmd *cobra.Command, args []string) {
			if *installHook {
				err := setup.InstallHook("")
				handleError(err)
				os.Exit(0)
			}
			if *uninstallHook {
				err := setup.UninstallHook()
				handleError(err)
				os.Exit(0)
			}
			if *showVersion {
				fmt.Println(versioninfo.Short())
				os.Exit(0)
			}

			err := cfg.Validate()
			if cmd.Flags().Changed("llm-provider") {
				cfg.LLMProvider = llmProvider
			}
			if err != nil || cfg.IsTestMode() || *runSetupWizard {
				newCfg, err := setup.Run(cfg)
				if err != nil {
					handleError(errors.Wrap(err, "failed to run setup wizard"))
				}
				if err := newCfg.Save(); err != nil {
					handleError(errors.Wrap(err, "failed to save new configuration"))
				}
				cfg = newCfg
			}

			if *runSetupWizard {
				fmt.Println(ui.Green.Bold(true).Render("SETTINGS SAVED."))
				os.Exit(0)
			}

			ctx := context.Background()

			provider, err := llmprovider.NewProvider(ctx, cfg)
			handleError(err)

			var commitMsgFile string
			if len(args) > 0 {
				commitMsgFile = args[0]
			}

			application := app.NewApp(provider, git.NewClient(*addAll), cfg.Prompt, cfg.IncludeProjectContext, cfg.RecentCommitsCount)
			err = application.Run(ctx, userMessage, hint, *yoloMode, *skipCI, commitMsgFile)
			if err != nil {
				handleError(err)
			}
		},
	}

	showVersion = rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	runSetupWizard = rootCmd.PersistentFlags().Bool("setup-wizard", false, "Run setup wizard")
	yoloMode = rootCmd.PersistentFlags().Bool("yolo", false, "Commit immediately without asking for confirmation")
	addAll = rootCmd.PersistentFlags().BoolP("all", "a", false, "Add all tracked files to the commit")
	skipCI = rootCmd.PersistentFlags().Bool("skip-ci", false, "Append [skip ci] to the commit message")
	rootCmd.PersistentFlags().StringVarP(&userMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().StringVar(&llmProvider, "llm-provider", cfg.LLMProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")
	rootCmd.PersistentFlags().StringVarP(&hint, "hint", "H", "", "Provide contextual guidance for the commit summary generation")
	installHook = rootCmd.PersistentFlags().Bool("install-hook", false, "Install git-commit-summary as a prepare-commit-msg hook")
	uninstallHook = rootCmd.PersistentFlags().Bool("uninstall-hook", false, "Uninstall git-commit-summary as a prepare-commit-msg hook")

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
