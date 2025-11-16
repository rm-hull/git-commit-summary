package config

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

//go:embed prompt.md
var prompt string

type Config struct {
	LLMProvider string `validate:"required,oneof=google openai llama.cpp"`
	APIKey      string `validate:"required"`
	Model       string `validate:"required"`
	BaseURL     string `validate:"required_if=LLMProvider llama.cpp"`
	Prompt      string
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
	}

	if cfg.LLMProvider == "" {
		cfg.LLMProvider = "google"
		cfg.Model = "gemini-2.5-flash-preview-09-2025"
	}

	switch cfg.LLMProvider {
	case "google":
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
		cfg.Model = os.Getenv("GEMINI_MODEL")
	case "openai":
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
		cfg.Model = os.Getenv("OPENAI_MODEL")
	case "llama.cpp":
		cfg.APIKey = os.Getenv("LLAMACPP_API_KEY")
		cfg.BaseURL = os.Getenv("LLAMACPP_BASE_URL")
		cfg.Model = os.Getenv("LLAMACPP_MODEL")
	}

	if cfg.LLMProvider == "google" && cfg.Model == "" {
		cfg.Model = "gemini-2.5-flash-preview-09-2025"
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (c *Config) Save() error {
	configPath, err := xdg.ConfigFile("git-commit-summary/config.env")
	if err != nil {
		return err
	}

	switch c.LLMProvider {
	case "google":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":   "google",
			"GEMINI_API_KEY": c.APIKey,
			"GEMINI_MODEL":   c.Model,
		})
	case "openai":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":   "openai",
			"OPENAI_API_KEY": c.APIKey,
			"OPENAI_MODEL":   c.Model,
		})
	case "llama.cpp":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":      "openai",
			"LLAMACPP_API_KEY":  c.APIKey,
			"LLAMACPP_MODEL":    c.Model,
			"LLAMACPP_BASE_URL": c.BaseURL,
		})
	default:
		return errors.New("unknown provider")
	}
}

func updateProperties(configPath string, props map[string]string) error {
	// Read existing file if present.
	original := []string{}
	data, err := os.ReadFile(configPath)
	if err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			original = append(original, scanner.Text())
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	updated := map[string]bool{}
	out := make([]string, 0, len(original))

	// Matches:
	//   KEY="VALUE"
	//   KEY=VALUE
	// Captures:
	//   1: KEY
	//   2: VALUE (inside quotes or unquoted)
	//   3: quote type = `"` or empty if unquoted
	re := regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)=(?:"(.*)"|(.*))\s*$`)

	for _, line := range original {
		m := re.FindStringSubmatch(line)
		if m == nil {
			// Keep raw/unknown lines
			out = append(out, line)
			continue
		}

		key := m[1]

		// m[2] is the quoted value (if present)
		// m[3] is the unquoted value (if present)
		quotedOriginal := m[2] != ""

		// Check if this key should be updated
		if newVal, ok := props[key]; ok {
			if quotedOriginal {
				out = append(out, fmt.Sprintf(`%s="%s"`, key, newVal))
			} else {
				out = append(out, fmt.Sprintf(`%s=%s`, key, newVal))
			}
			updated[key] = true
		} else {
			out = append(out, line)
		}
	}

	// Append any keys in props missing from the file (default to quoted form)
	for k, v := range props {
		if !updated[k] {
			out = append(out, fmt.Sprintf(`%s="%s"`, k, v))
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	// Write new content
	buf := strings.Join(out, "\n") + "\n"
	return os.WriteFile(configPath, []byte(buf), 0o644)
}
