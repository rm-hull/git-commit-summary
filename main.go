package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/rm-hull/git-commit-summary/internal/app"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/rm-hull/git-commit-summary/internal/help"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/setup"
	"github.com/rm-hull/git-commit-summary/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed HELP.md
var helpText string

func main() {
	if os.Getenv("GIT_COMMIT_SUMMARY_IGNORE_HOOK") == "1" {
		os.Exit(0)
	}

	runOpts := app.RunOptions{}

	cfg, err := config.Load()
	runOpts.HandleError(err)

	var llmProvider string
	var runSetupWizard *bool
	var showVersion *bool
	var addAll *bool
	var genBash *bool
	var genZsh *bool
	var installHook *bool
	var uninstallHook *bool
	var maxTokens int

	rootCmd := &cobra.Command{
		Use:   "git-commit-summary",
		Short: fmt.Sprintf("Generate a commit summary using Gemini, OpenAI, Llama.cpp, OpenRouter (version: %s)", versioninfo.Short()),
		Long:  help.RenderMarkdownHelp(helpText, help.TerminalWidth()),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				runOpts.CommitMsgFile = args[0]
			}
			if *installHook {
				err := setup.InstallHook()
				runOpts.HandleError(err)
				os.Exit(0)
			}
			if *uninstallHook {
				err := setup.UninstallHook()
				runOpts.HandleError(err)
				os.Exit(0)
			}
			if *genBash {
				_ = cmd.GenBashCompletion(os.Stdout)
				os.Exit(0)
			}
			if *genZsh {
				_ = cmd.GenZshCompletion(os.Stdout)
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
					runOpts.HandleError(errors.Wrap(err, "failed to run setup wizard"))
				}
				if err := newCfg.Save(); err != nil {
					runOpts.HandleError(errors.Wrap(err, "failed to save new configuration"))
				}
				cfg = newCfg
			}

			if *runSetupWizard {
				fmt.Println(ui.Green.Bold(true).Render("SETTINGS SAVED."))
				os.Exit(0)
			}

			ctx := context.Background()

			provider, err := llmprovider.NewProvider(ctx, cfg)
			if errors.Is(err, interfaces.ErrAborted) && runOpts.CommitMsgFile != "" {
				fmt.Println(ui.BoldRed.Render("ABORTED!"))
				os.Exit(1)
			}
			runOpts.HandleError(err)

			// Use config default, but allow CLI flag to override
			finalMaxTokens := cfg.MaxTokens
			if cmd.Flags().Changed("max-tokens") {
				finalMaxTokens = maxTokens
			}
			application := app.NewApp(provider, git.NewClient(*addAll, finalMaxTokens), cfg.Prompt, cfg.IncludeProjectContext, cfg.RecentCommitsCount)
			err = application.Run(ctx, runOpts)
			runOpts.HandleError(err)
		},
	}

	showVersion = rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
	genBash = rootCmd.PersistentFlags().Bool("bash", false, "Generate bash completion script")
	genZsh = rootCmd.PersistentFlags().Bool("zsh", false, "Generate zsh completion script")
	runSetupWizard = rootCmd.PersistentFlags().Bool("setup-wizard", false, "Run setup wizard")
	addAll = rootCmd.PersistentFlags().BoolP("all", "a", false, "Add all tracked files to the commit")
	rootCmd.PersistentFlags().BoolVar(&runOpts.Yolo, "yolo", false, "Commit immediately without asking for confirmation")
	rootCmd.PersistentFlags().BoolVar(&runOpts.SkipCI, "skip-ci", false, "Append [skip ci] to the commit message")
	rootCmd.PersistentFlags().BoolVar(&runOpts.NoVerify, "no-verify", false, "Bypass pre-commit and commit-msg hooks")
	rootCmd.PersistentFlags().StringVarP(&runOpts.UserMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().StringVarP(&runOpts.Hint, "hint", "H", "", "Provide contextual guidance for the commit summary generation")
	rootCmd.PersistentFlags().StringVar(&llmProvider, "llm-provider", cfg.LLMProvider, "Use specific LLM provider, overrides environment variable LLM_PROVIDER")
	installHook = rootCmd.PersistentFlags().Bool("install-hook", false, "Install git-commit-summary as a prepare-commit-msg hook")
	uninstallHook = rootCmd.PersistentFlags().Bool("uninstall-hook", false, "Uninstall git-commit-summary as a prepare-commit-msg hook")
	rootCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", 0, "Maximum number of tokens to spend for per-file diff processing")

	_ = rootCmd.Execute()
}
