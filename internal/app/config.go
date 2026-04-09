package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/celiumgrid/git-digest/internal/ai"
	"github.com/celiumgrid/git-digest/internal/i18n"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

// Config is the unified runtime configuration.
type Config struct {
	Language string         `json:"language,omitempty"`
	Time     timequery.Spec `json:"time,omitempty"`

	Format     string `json:"format,omitempty"`
	OutputFile string `json:"output,omitempty"`

	RepoPath  string `json:"repo,omitempty"`
	ReposPath string `json:"repos,omitempty"`

	Provider string `json:"provider,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	Model    string `json:"model,omitempty"`
	Author   string `json:"author,omitempty"`
	Prompt   string `json:"prompt,omitempty"`

	Interactive bool   `json:"interactive,omitempty"`
	ConfigPath  string `json:"-"`
	NoConfig    bool   `json:"-"`

	SaveAsDefault bool `json:"-"`
}

func DefaultConfig() Config {
	return Config{
		Language: i18n.LanguageEnglish,
		Time:     timequery.DefaultSpec(),
		Format:   "text",
		Prompt:   "basic",
		Provider: ai.DefaultProvider,
		BaseURL:  ai.DefaultBaseURL(ai.DefaultProvider),
		Model:    ai.DefaultModelName(ai.DefaultProvider),
	}
}

func ShouldUseInteractive(args []string, interactive bool) bool {
	return interactive || len(args) == 0
}

func MergeConfig(base, fileCfg, cliCfg Config, changed map[string]bool) Config {
	merged := base
	applyNonEmpty(&merged, fileCfg)

	if changed["language"] {
		merged.Language = cliCfg.Language
	}
	if changed["period"] || changed["on"] || changed["from"] || changed["to"] {
		merged.Time = cliCfg.Time
	}
	if changed["format"] {
		merged.Format = cliCfg.Format
	}
	if changed["output"] {
		merged.OutputFile = cliCfg.OutputFile
	}
	if changed["repo"] {
		merged.RepoPath = cliCfg.RepoPath
	}
	if changed["repos"] {
		merged.ReposPath = cliCfg.ReposPath
	}
	if changed["provider"] {
		merged.Provider = cliCfg.Provider
	}
	if changed["base-url"] {
		merged.BaseURL = cliCfg.BaseURL
	}
	if changed["api-key"] {
		merged.APIKey = cliCfg.APIKey
	}
	if changed["model"] {
		merged.Model = cliCfg.Model
	}
	if changed["author"] {
		merged.Author = cliCfg.Author
	}
	if changed["prompt"] {
		merged.Prompt = cliCfg.Prompt
	}
	if changed["interactive"] {
		merged.Interactive = cliCfg.Interactive
	}

	merged.Language = i18n.NormalizeLanguage(merged.Language)
	if !timequery.HasValue(merged.Time) {
		merged.Time = timequery.DefaultSpec()
	}
	if merged.Provider == "" {
		merged.Provider = ai.DefaultProvider
	}
	if merged.BaseURL == "" {
		merged.BaseURL = ai.DefaultBaseURL(merged.Provider)
	}
	if merged.Model == "" {
		merged.Model = ai.DefaultModelName(merged.Provider)
	}
	if merged.Prompt == "" {
		merged.Prompt = "basic"
	}
	if merged.Format == "" {
		merged.Format = "text"
	}

	return merged
}

func applyNonEmpty(dst *Config, src Config) {
	if src.Language != "" {
		dst.Language = src.Language
	}
	if timequery.HasValue(src.Time) {
		dst.Time = src.Time
	}
	if src.Format != "" {
		dst.Format = src.Format
	}
	if src.OutputFile != "" {
		dst.OutputFile = src.OutputFile
	}
	if src.RepoPath != "" {
		dst.RepoPath = src.RepoPath
	}
	if src.ReposPath != "" {
		dst.ReposPath = src.ReposPath
	}
	if src.Provider != "" {
		dst.Provider = src.Provider
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.APIKey != "" {
		dst.APIKey = src.APIKey
	}
	if src.Model != "" {
		dst.Model = src.Model
	}
	if src.Author != "" {
		dst.Author = src.Author
	}
	if src.Prompt != "" {
		dst.Prompt = src.Prompt
	}
	if src.Interactive {
		dst.Interactive = true
	}
}

func ValidateConfig(cfg Config) error {
	if cfg.RepoPath != "" && cfg.ReposPath != "" {
		if cfg.Language == timequery.LanguageChinese {
			return errors.New(i18n.T(cfg.Language, "config.repo_conflict"))
		}
		return errors.New(i18n.T(cfg.Language, "config.repo_conflict"))
	}

	if cfg.Format != "text" && cfg.Format != "markdown" {
		return fmt.Errorf(i18n.T(cfg.Language, "config.format"), cfg.Format)
	}

	_, err := timequery.ResolveWithLanguage(cfg.Time, nil, nowForValidation(), cfg.Language)
	return err
}

var nowForValidation = time.Now

func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf(i18n.T(i18n.LanguageEnglish, "config.user_config_dir"), err)
	}
	return filepath.Join(configDir, "git-digest", "config.json"), nil
}
