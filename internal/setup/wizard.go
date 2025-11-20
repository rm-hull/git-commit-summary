package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
)

func Run(cfg *config.Config) (*config.Config, error) {

	var confirm bool

	options := []huh.Option[string]{
		huh.NewOption("Google (Gemini)", "google"),
		huh.NewOption("OpenAI", "openai"),
		huh.NewOption("Llama.cpp", "llama.cpp"),
	}

	if cfg.IsTestMode() {
		options = append(options, huh.NewOption("Test", "test"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select LLM Provider").
				Options(options...).
				Value(&cfg.LLMProvider),
		),
		geminiGroup(cfg),
		openaiGroup(cfg),
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
			Options(
				huh.NewOption("Gemini 3 Pro (Preview)", "gemini-3-pro-preview"),
				huh.NewOption("Gemini 2.5 Pro", "gemini-2.5-pro"),
				huh.NewOption("Gemini 2.5 Flash", "gemini-2.5-flash"),
				huh.NewOption("Gemini 2.5 Flash-Lite", "gemini-2.5-flash-lite"),
				huh.NewOption("Gemini 2.5 Flash (Preview 09-2025)", "gemini-2.5-flash-preview-09-2025"),
				huh.NewOption("Gemini Flash (latest)", "gemini-flash-latest"),
			),
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
			Options(
				huh.NewOption("GPT 5.1", "gpt-5.1"),
				huh.NewOption("GPT 5.1 mini", "gpt-5.1-mini"),
				huh.NewOption("GPT 5.0 nano", "gpt-5.0-nano"),
				huh.NewOption("GPT 4.1", "gpt-4.1"),
				huh.NewOption("ChatGPT 4o latest", "chatgpt-4o-latest"),
				huh.NewOption("o3", "o3"),
			),
		huh.NewInput().
			Title("API Key").
			Value(&cfg.APIKey),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "openai"
	})
}

func llamacppGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("Llama.CPP Model").
			Value(&cfg.Model).
			Options(
				huh.NewOption("Qwen2.5-Coder-7B-Instruct-Q4_K_M", "Qwen2.5-Coder-7B-Instruct-Q4_K_M"),
				huh.NewOption("DeepSeek-R1-Distill-Qwen-7B-Q6_K", "DeepSeek-R1-Distill-Qwen-7B-Q6_K"),
				huh.NewOption("Meta-Llama-3.1-8B-Instruct-Q6_K", "Meta-Llama-3.1-8B-Instruct-Q6_K"),
				huh.NewOption("Meta-Llama-3.1-8B-Instruct-Q4_K_M", "Meta-Llama-3.1-8B-Instruct-Q4_K_M"),
				huh.NewOption("Phi-4-mini-instruct-Q4_K_M", "Phi-4-mini-instruct-Q4_K_M"),
				huh.NewOption("gemma-3-12b-it-Q4_K_M", "gemma-3-12b-it-Q4_K_M"),
				huh.NewOption("gemma-3-4b-it-Q8_0", "gemma-3-4b-it-Q8_0"),
				huh.NewOption("gemma-3-12b-it-q4_0", "gemma-3-12b-it-q4_0"),
				huh.NewOption("gemma-3-4b-it-q4_0", "gemma-3-4b-it-q4_0"),
			),
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
				return fmt.Sprintf(
					"Using \"%s\" provider with:\n  - API Key (%s)\n  - Model (%s)",
					cfg.LLMProvider, cfg.APIKey, cfg.Model)
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
