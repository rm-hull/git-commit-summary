package setup

import (
	"github.com/charmbracelet/huh"
	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
)

func Run(cfg *config.Config) (*config.Config, error) {

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
	)

	err := form.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, interfaces.ErrAborted
		} else {
			return nil, err
		}
	}

	return cfg, nil
}

func geminiGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("Gemini API Key").
			Value(&cfg.APIKey),
		huh.NewSelect[string]().
			Title("Model").
			Value(&cfg.Model).
			Options(
				huh.NewOption("Gemini 2.5 Pro", "gemini-2.5-pro"),
				huh.NewOption("Gemini 2.5 Flash", "gemini-2.5-flash"),
				huh.NewOption("Gemini 2.5 Flash-Lite", "gemini-2.5-flash-lite"),
				huh.NewOption("Gemini 2.5 Flash (Preview 09-2025)", "gemini-2.5-flash-preview-09-2025"),
				huh.NewOption("Gemini Flash (latest)", "gemini-flash-latest"),
			),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "google"
	})
}

func openaiGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("OpenAI API Key").
			Value(&cfg.APIKey),
		huh.NewSelect[string]().
			Title("Model").
			Value(&cfg.Model).
			Options(
				huh.NewOption("GPT 5.1", "gpt-5.1"),
				huh.NewOption("GPT 5.1 mini", "gpt-5.1-mini"),
				huh.NewOption("GPT 5.0 nano", "gpt-5.0-nano"),
				huh.NewOption("GPT 4.1", "gpt-4.1"),
				huh.NewOption("ChatGPT 4o latest", "chatgpt-4o-latest"),
				huh.NewOption("o3", "o3"),
			),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "openai"
	})
}

func llamacppGroup(cfg *config.Config) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("Llama.cpp API Key").
			Value(&cfg.APIKey),
		huh.NewInput().
			Title("Model").
			Value(&cfg.Model).
			Placeholder("gemma3"),
		huh.NewInput().
			Title("Base URL").
			Value(&cfg.BaseURL),
	).WithHideFunc(func() bool {
		return cfg.LLMProvider != "llama.cpp"
	})
}
