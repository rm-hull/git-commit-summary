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
	if os.Getenv("GIT_COMMIT_SUMMARY_IGNORE_HOOK") == "1" {
		os.Exit(0)
	}

	cfg, err := config.Load()
	handleError(err)

	var llmProvider string
	var runSetupWizard *bool
	var showVersion *bool
	var addAll *bool
	var installHook *bool
	var uninstallHook *bool

	runOptions := app.RunOptions{}

	rootCmd := &cobra.Command{
		Use:   "git-commit-summary",
		Short: fmt.Sprintf("Generate a commit summary using Gemini, OpenAI, Llama.cpp, OpenRouter (version: %s)", versioninfo.Short()),
		Run: func(cmd *cobra.Command, args []string) {
			if *installHook {
				err := setup.InstallHook()
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
			if errors.Is(err, interfaces.ErrAborted) && runOptions.CommitMsgFile != "" {
				fmt.Println(ui.BoldRed.Render("ABORTED!"))
				os.Exit(1)
			}
			handleError(err)

			if len(args) > 0 {
				runOptions.CommitMsgFile = args[0]
			}

			application := app.NewApp(provider, git.NewClient(*addAll), cfg.Prompt, cfg.IncludeProjectContext, cfg.RecentCommitsCount)
			err = application.Run(ctx, runOptions)
			handleError(err)
		},
	}

	showVersion = rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	runSetupWizard = rootCmd.PersistentFlags().Bool("setup-wizard", false, "Run setup wizard")
	addAll = rootCmd.PersistentFlags().BoolP("all", "a", false, "Add all tracked files to the commit")
	rootCmd.PersistentFlags().BoolVar(&runOptions.Yolo, "yolo", false, "Commit immediately without asking for confirmation")
	rootCmd.PersistentFlags().BoolVar(&runOptions.SkipCI, "skip-ci", false, "Append [skip ci] to the commit message")
	rootCmd.PersistentFlags().BoolVar(&runOptions.NoVerify, "no-verify", false, "Bypass pre-commit and commit-msg hooks")
	rootCmd.PersistentFlags().StringVarP(&runOptions.UserMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().StringVarP(&runOptions.Hint, "hint", "H", "", "Provide contextual guidance for the commit summary generation")
	rootCmd.PersistentFlags().StringVar(&llmProvider, "llm-provider", cfg.LLMProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")
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
