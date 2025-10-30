package config

import (
	_ "embed"
	"os"

	"github.com/adrg/xdg"
	"github.com/joho/godotenv"
)

//go:embed prompt.md
var prompt string

type GeminiConfig struct {
	APIKey string
	Model  string
}

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type Config struct {
	LLMProvider string
	Prompt      string
	Gemini      GeminiConfig
	OpenAI      OpenAIConfig
}

func Load() (*Config, error) {
	// Load XDG config file
	configFile, err := xdg.ConfigFile("git-commit-summary/config.env")
	if err != nil {
		return nil, err
	}
	_ = godotenv.Load(configFile)

	// Load local .env file, overriding XDG config
	_ = godotenv.Overload(".env")

	cfg := &Config{
		LLMProvider: os.Getenv("LLM_PROVIDER"),
		Prompt:      prompt,
		Gemini: GeminiConfig{
			APIKey: os.Getenv("GEMINI_API_KEY"),
			Model:  os.Getenv("GEMINI_MODEL"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
			Model:   os.Getenv("OPENAI_MODEL"),
		},
	}

	if cfg.LLMProvider == "" {
		cfg.LLMProvider = "google"
	}

	if cfg.Gemini.Model == "" {
		cfg.Gemini.Model = "gemini-2.5-flash-preview-09-2025"
	}

	if cfg.OpenAI.Model == "" {
		cfg.OpenAI.Model = "gpt-4o"
	}

	return cfg, nil
}
