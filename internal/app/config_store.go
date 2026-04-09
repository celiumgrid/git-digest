package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kway-teow/git-digest/internal/i18n"
)

func LoadConfig(path, language string) (Config, error) {
	if path == "" {
		return Config{}, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf(i18n.T(language, "config_store.read"), err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf(i18n.T(language, "config_store.parse"), err)
	}

	return cfg, nil
}

func SaveConfig(path string, cfg Config, language string) error {
	if path == "" {
		return errors.New(i18n.T(language, "config_store.path_empty"))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.mkdir"), err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.marshal"), err)
	}

	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf(i18n.T(language, "config_store.write"), err)
	}
	return nil
}
