package config

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

//go:embed prompt.md
var prompt string

//go:embed models.json
var models_raw []byte

type Models struct {
	LastUpdated *time.Time `json:"last_updated"`
	Providers   map[string][]struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Model       string `json:"model"`
		Deprecated  bool   `json:"deprecated"`
	} `json:"providers"`
}

type Config struct {
	LLMProvider           string `validate:"required,oneof=google openai llama.cpp openrouter test"`
	APIKey                string `validate:"required_unless=LLMProvider test"`
	Model                 string `validate:"required_unless=LLMProvider test"`
	BaseURL               string `validate:"required_if=LLMProvider llama.cpp"`
	IncludeProjectContext bool
	Models                Models
	Prompt                string
	validate              *validator.Validate
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
		validate:    validator.New(),
	}
	// Default to true if not specified.
	cfg.IncludeProjectContext = true
	if os.Getenv("INCLUDE_PROJECT_CONTEXT") == "false" {
		cfg.IncludeProjectContext = false
	}

	err = json.Unmarshal(models_raw, &cfg.Models)
	if err != nil {
		return nil, err
	}

	if cfg.LLMProvider == "" {
		cfg.LLMProvider = "google"
	}

	switch cfg.LLMProvider {
	case "google":
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
		cfg.Model = os.Getenv("GEMINI_MODEL")
		if cfg.Model == "" {
			cfg.Model = "gemini-3-flash-preview"
		}
	case "openai":
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
		cfg.Model = os.Getenv("OPENAI_MODEL")
		if cfg.Model == "" {
			cfg.Model = "gpt-4o"
		}
	case "openrouter":
		cfg.APIKey = os.Getenv("OPENROUTER_API_KEY")
		cfg.Model = os.Getenv("OPENROUTER_MODEL")
		if cfg.Model == "" {
			cfg.Model = "qwen/qwen3-coder:free"
		}
	case "llama.cpp":
		cfg.APIKey = os.Getenv("LLAMACPP_API_KEY")
		cfg.BaseURL = os.Getenv("LLAMACPP_BASE_URL")
		cfg.Model = os.Getenv("LLAMACPP_MODEL")
	}

	return cfg, nil
}

func (cfg *Config) IsTestMode() bool {
	return cfg.LLMProvider == "test"
}

func (cfg *Config) Validate() error {
	return cfg.validate.Struct(cfg)
}

func (cfg *Config) Save() error {
	configPath, err := xdg.ConfigFile("git-commit-summary/config.env")
	if err != nil {
		return err
	}

	switch cfg.LLMProvider {
	case "google":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":            "google",
			"GEMINI_API_KEY":          cfg.APIKey,
			"GEMINI_MODEL":            cfg.Model,
			"INCLUDE_PROJECT_CONTEXT": fmt.Sprintf("%t", cfg.IncludeProjectContext),
		})
	case "openai":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":            "openai",
			"OPENAI_API_KEY":          cfg.APIKey,
			"OPENAI_MODEL":            cfg.Model,
			"INCLUDE_PROJECT_CONTEXT": fmt.Sprintf("%t", cfg.IncludeProjectContext),
		})
	case "openrouter":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":            "openrouter",
			"OPENROUTER_API_KEY":      cfg.APIKey,
			"OPENROUTER_MODEL":        cfg.Model,
			"INCLUDE_PROJECT_CONTEXT": fmt.Sprintf("%t", cfg.IncludeProjectContext),
		})
	case "llama.cpp":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":            "llama.cpp",
			"LLAMACPP_API_KEY":        cfg.APIKey,
			"LLAMACPP_MODEL":          cfg.Model,
			"LLAMACPP_BASE_URL":       cfg.BaseURL,
			"INCLUDE_PROJECT_CONTEXT": fmt.Sprintf("%t", cfg.IncludeProjectContext),
		})
	case "test":
		return updateProperties(configPath, map[string]string{
			"LLM_PROVIDER":            "test",
			"INCLUDE_PROJECT_CONTEXT": fmt.Sprintf("%t", cfg.IncludeProjectContext),
		})
	default:
		return errors.Newf("unknown LLM provider: %s", cfg.LLMProvider)
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
	//   2: VALUE (raw, potentially quoted)
	re := regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

	for _, line := range original {
		m := re.FindStringSubmatch(line)
		if m == nil {
			// Keep raw/unknown lines
			out = append(out, line)
			continue
		}

		key := m[1]
		val := strings.TrimSpace(m[2])
		oldVal := val
		if len(val) >= 2 && strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			oldVal = val[1 : len(val)-1]
		}

		// Check if this key should be updated
		if newVal, ok := props[key]; ok {
			if oldVal != newVal {
				out = append(out, "# "+line)
			}
			out = append(out, fmt.Sprintf(`%s="%s"`, key, newVal))
			updated[key] = true
		} else {
			out = append(out, line)
		}
	}

	// Append any keys in props missing from the file (default to quoted form)
	var newKeys []string
	for k := range props {
		if !updated[k] {
			newKeys = append(newKeys, k)
		}
	}
	sort.Strings(newKeys)

	for _, k := range newKeys {
		out = append(out, fmt.Sprintf(`%s="%s"`, k, props[k]))
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	// Write new content
	buf := strings.Join(out, "\n") + "\n"
	return os.WriteFile(configPath, []byte(buf), 0o600)
}
