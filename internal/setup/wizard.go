package setup

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

func Run(cfg *config.Config) (*config.Config, error) {

	var confirm bool

	options := []huh.Option[string]{
		huh.NewOption("Google (Gemini)", "google"),
		huh.NewOption("OpenAI", "openai"),
		huh.NewOption("OpenRouter", "openrouter"),
		huh.NewOption("Llama.cpp", "llama.cpp"),
	}

	if cfg.IsTestMode() {
		options = append(options, huh.NewOption("Test", "test"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include project context in prompts?").
				Description("If enabled, the tool will include your README.md or\n.project-context.md to help the LLM understand your\nproject better.").
				Affirmative("Yes").
				Negative("No").
				Value(&cfg.IncludeProjectContext),
			huh.NewSelect[int]().
				Title("Number of recent commits to include").
				Description("The number of recent commits to include as context to help\nthe LLM understand the current trajectory of the project.").
				Options(
					huh.NewOption("0", 0),
					huh.NewOption("1", 1),
					huh.NewOption("3", 3),
					huh.NewOption("5", 5),
				).
				Value(&cfg.RecentCommitsCount),
		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select LLM Provider").
				Options(options...).
				Value(&cfg.LLMProvider),
		),
		geminiGroup(cfg),
		openaiGroup(cfg),
		openrouterGroup(cfg),
		llamacppGroup(cfg),
		validationGroup(cfg),
		submitGroup(cfg, &confirm),
	)

	err := form.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, interfaces.ErrAborted
		} else {
			return nil, err
		}
	}

	if !confirm {
		return nil, interfaces.ErrAborted
	}

	return cfg, nil
}

func geminiGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("Google Model").
			Value(&cfg.Model).
			Options(options(cfg, "google")...),
		huh.NewInput().
			Title("API Key").
			Value(&cfg.APIKey),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "google"
	})
}

func openaiGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("OpenAI Model").
			Value(&cfg.Model).
			Options(options(cfg, "openai")...),
		huh.NewInput().
			Title("API Key").
			Value(&cfg.APIKey),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "openai"
	})
}

func openrouterGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("OpenRouter Model").
			Value(&cfg.Model).
			Options(options(cfg, "openrouter")...),
		huh.NewInput().
			Title("API Key").
			Value(&cfg.APIKey),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "openrouter"
	})
}

func llamacppGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("Llama.CPP Model").
			Value(&cfg.Model).
			Options(options(cfg, "llama.cpp")...),
		huh.NewInput().
			Title("API Key").
			Value(&cfg.APIKey),
		huh.NewInput().
			Title("Base URL").
			Value(&cfg.BaseURL),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "llama.cpp"
	})
}

func submitGroup(cfg *config.Config, confirm *bool) *huh.Group {
	return huh.NewGroup(
		huh.NewConfirm().
			Title("Confirm overwrite settings?").
			Affirmative("Yes").
			Negative("No").
			Value(confirm).
			DescriptionFunc(func() string {
				contextStr := "NO"
				if cfg.IncludeProjectContext {
					contextStr = "YES"
				}
				return fmt.Sprintf(
					"Using \"%s\" provider with:\n  - API Key (%s)\n  - Model (%s)\n  - Include project context (%s)\n  - Recent commits count (%d)",
					cfg.LLMProvider, cfg.APIKey, cfg.Model, contextStr, cfg.RecentCommitsCount)

			}, cfg),
	).WithHideFunc(func() bool {
		return cfg.Validate() != nil
	})
}

func validationGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewNote().
			Title("The following fields are required:").
			DescriptionFunc(func() string {
				parts := make([]string, 0)
				if err := cfg.Validate(); err != nil {
					if ve, ok := err.(validator.ValidationErrors); ok {
						for _, err := range ve {
							parts = append(parts, fmt.Sprintf("  - %s", err.Field()))
						}
					}
				}
				return fmt.Sprintf("%s\n\nGo back and correct these issues", strings.Join(parts, "\n"))
			}, cfg),
	).WithHideFunc(func() bool {
		return cfg.Validate() == nil
	})
}

func options(cfg *config.Config, provider string) []huh.Option[string] {

	options := make([]huh.Option[string], 0, 10)

	for _, opt := range cfg.Models.Providers[provider] {
		name := opt.Name
		if name == "" {
			name = opt.Model
		}

		if opt.Deprecated {
			name = ui.Strikethrough.Render(name) + " (deprecated)"
		}
		options = append(options, huh.NewOption(name, opt.Model))
	}

	return options
}
