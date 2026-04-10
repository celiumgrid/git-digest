package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/celiumgrid/git-digest/internal/i18n"
	"github.com/celiumgrid/git-digest/internal/pathutil"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

type storedConfig struct {
	Language    string          `json:"language,omitempty"`
	Time        *timequery.Spec `json:"time,omitempty"`
	Format      string          `json:"format,omitempty"`
	OutputFile  string          `json:"output,omitempty"`
	RepoPath    string          `json:"repo,omitempty"`
	ReposPath   string          `json:"repos,omitempty"`
	Provider    string          `json:"provider,omitempty"`
	BaseURL     string          `json:"base_url,omitempty"`
	APIKey      string          `json:"api_key,omitempty"`
	Model       string          `json:"model,omitempty"`
	Author      string          `json:"author,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	Interactive bool            `json:"interactive,omitempty"`
}

func LoadConfig(path, language string) (Config, error) {
	if path == "" {
		return Config{}, nil
	}

	var err error
	path, err = pathutil.NormalizeUserPath(path)
	if err != nil {
		return Config{}, fmt.Errorf(i18n.T(language, "config_store.read"), err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf(i18n.T(language, "config_store.read"), err)
	}

	var stored storedConfig
	if err := json.Unmarshal(b, &stored); err != nil {
		return Config{}, fmt.Errorf(i18n.T(language, "config_store.parse"), err)
	}

	return configFromStored(stored), nil
}

func SaveConfig(path string, cfg Config, language string) error {
	if path == "" {
		return errors.New(i18n.T(language, "config_store.path_empty"))
	}

	var err error
	path, err = pathutil.NormalizeUserPath(path)
	if err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.write"), err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.mkdir"), err)
	}

	//nolint:gosec // The config file intentionally persists the user-supplied API key.
	b, err := json.MarshalIndent(storedFromConfig(cfg), "", "  ")
	if err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.marshal"), err)
	}

	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.write"), err)
	}
	return nil
}

func storedFromConfig(cfg Config) storedConfig {
	stored := storedConfig{
		Language:    cfg.Language,
		Format:      cfg.Format,
		OutputFile:  cfg.OutputFile,
		RepoPath:    cfg.RepoPath,
		ReposPath:   cfg.ReposPath,
		Provider:    cfg.Provider,
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		Author:      cfg.Author,
		Prompt:      cfg.Prompt,
		Interactive: cfg.Interactive,
	}
	if timequery.HasValue(cfg.Time) {
		timeCopy := cfg.Time
		stored.Time = &timeCopy
	}
	return stored
}

func configFromStored(stored storedConfig) Config {
	cfg := Config{
		Language:    stored.Language,
		Format:      stored.Format,
		OutputFile:  stored.OutputFile,
		RepoPath:    stored.RepoPath,
		ReposPath:   stored.ReposPath,
		Provider:    stored.Provider,
		BaseURL:     stored.BaseURL,
		APIKey:      stored.APIKey,
		Model:       stored.Model,
		Author:      stored.Author,
		Prompt:      stored.Prompt,
		Interactive: stored.Interactive,
	}
	if stored.Time != nil {
		cfg.Time = *stored.Time
	}
	return cfg
}
