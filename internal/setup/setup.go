package setup

import (
	"github.com/charmbracelet/huh"
	"github.com/rm-hull/git-commit-summary/internal/config"
)

func Run() (*config.Config, error) {
	cfg := &config.Config{
		LLMProvider: "google", // Default provider
		Gemini:      config.GeminiConfig{},
		OpenAI:      config.OpenAIConfig{},
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select LLM Provider").
				Options(
					huh.NewOption("Google (Gemini)", "google"),
					huh.NewOption("OpenAI", "openai"),
				).
				Value(&cfg.LLMProvider),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Gemini API Key").
				Value(&cfg.Gemini.APIKey).
				EchoMode(huh.EchoModePassword),
			huh.NewSelect[string]().
				Title("Gemini Model").
				Value(&cfg.Gemini.Model).
				Options(
					huh.NewOption("Gemini 2.5 Pro", "gemini-2.5-pro"),
					huh.NewOption("Gemini 2.5 Flash", "gemini-2.5-flash"),
					huh.NewOption("Gemini 2.5 Flash-Lite", "gemini-2.5-flash-lite"),
					huh.NewOption("Gemini 2.5 Flash (Preview 09-2025)", "gemini-2.5-flash-preview-09-2025"),
					huh.NewOption("Gemini Flash (latest)", "gemini-flash-latest"),
				),
		).WithHideFunc(func() bool {
			return cfg.LLMProvider != "google"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("OpenAI API Key").
				Value(&cfg.OpenAI.APIKey).
				EchoMode(huh.EchoModePassword),
			huh.NewInput().
				Title("OpenAI Model").
				Value(&cfg.OpenAI.Model).
				Placeholder("gpt-4o"),
			huh.NewInput().
				Title("OpenAI Base URL (optional)").
				Value(&cfg.OpenAI.BaseURL),
		).WithHideFunc(func() bool {
			return cfg.LLMProvider != "openai"
		}),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Set default values if they are empty
	if cfg.LLMProvider == "google" && cfg.Gemini.Model == "" {
		cfg.Gemini.Model = "gemini-2.5-flash-preview-09-2025"
	}
	if cfg.LLMProvider == "openai" && cfg.OpenAI.Model == "" {
		cfg.OpenAI.Model = "gpt-4o"
	}

	return cfg, nil
}
