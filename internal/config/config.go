package config

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

//go:embed prompt.md
var prompt string

type GeminiConfig struct {
	APIKey string `validate:"required"`
	Model  string
}

type OpenAIConfig struct {
	APIKey  string `validate:"required"`
	Model   string
	BaseURL string
}

type Config struct {
	LLMProvider string `validate:"required,oneof=google openai"`
	Prompt      string
	Gemini      GeminiConfig `validate:"required_if=LLMProvider google"`
	OpenAI      OpenAIConfig `validate:"required_if=LLMProvider openai"`
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

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (c *Config) Save() error {
	configFile, err := xdg.ConfigFile("git-commit-summary/config.env")
	if err != nil {
		return err
	}

	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if c.LLMProvider == "google" {
		_, err = file.WriteString(fmt.Sprintf("LLM_PROVIDER=google\nGEMINI_API_KEY=%s\nGEMINI_MODEL=%s\n", c.Gemini.APIKey, c.Gemini.Model))
	} else {
		_, err = file.WriteString(fmt.Sprintf("LLM_PROVIDER=openai\nOPENAI_API_KEY=%s\nOPENAI_MODEL=%s\nOPENAI_BASE_URL=%s\n", c.OpenAI.APIKey, c.OpenAI.Model, c.OpenAI.BaseURL))
	}

	return err
}
