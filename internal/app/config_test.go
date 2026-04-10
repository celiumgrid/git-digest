package app

import (
	"path/filepath"
	"testing"

	"github.com/celiumgrid/git-digest/internal/ai"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

func TestShouldUseInteractive(t *testing.T) {
	if !ShouldUseInteractive(nil, false) {
		t.Fatalf("expected interactive mode when no args")
	}
	if ShouldUseInteractive([]string{"--period", "last-7d"}, false) {
		t.Fatalf("expected non-interactive mode when args are provided")
	}
	if !ShouldUseInteractive([]string{"--period", "last-7d"}, true) {
		t.Fatalf("expected interactive mode when interactive flag is set")
	}
}

func TestDefaultConfigIncludesAIProviderDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Language != timequery.LanguageEnglish {
		t.Fatalf("language default mismatch: got %q", cfg.Language)
	}
	if cfg.Provider != ai.DefaultProvider {
		t.Fatalf("provider default mismatch: got %q want %q", cfg.Provider, ai.DefaultProvider)
	}
	if cfg.BaseURL != ai.DefaultBaseURL(ai.DefaultProvider) {
		t.Fatalf("base url default mismatch: got %q", cfg.BaseURL)
	}
	if cfg.Model != ai.DefaultModelName(ai.DefaultProvider) {
		t.Fatalf("model default mismatch: got %q", cfg.Model)
	}
}

func TestMergeConfigPriority(t *testing.T) {
	base := DefaultConfig()
	fileCfg := Config{
		Language: timequery.LanguageChinese,
		Time:     timequery.Spec{Kind: timequery.KindSingleDay, On: "2026-04-01"},
		Format:   "markdown",
		RepoPath: "/repos/a",
		Provider: ai.ProviderOpenAI,
		BaseURL:  "https://example.com/v1",
		APIKey:   "file-key",
		Model:    "file-model",
	}
	cli := Config{
		Language:   timequery.LanguageEnglish,
		Time:       timequery.Spec{Kind: timequery.KindRange, From: "2026-04-01", To: "2026-04-09"},
		OutputFile: "out.md",
		Provider:   ai.ProviderDeepSeek,
		APIKey:     "cli-key",
	}
	changed := map[string]bool{
		"language": true,
		"from":     true,
		"to":       true,
		"output":   true,
		"provider": true,
		"api-key":  true,
	}

	merged := MergeConfig(base, fileCfg, cli, changed)
	if merged.Language != timequery.LanguageEnglish {
		t.Fatalf("language should be overridden by cli, got %q", merged.Language)
	}
	if merged.Time.Kind != timequery.KindRange {
		t.Fatalf("time should be overridden by cli, got %+v", merged.Time)
	}
	if merged.Format != "markdown" {
		t.Fatalf("format should come from config file, got %q", merged.Format)
	}
	if merged.RepoPath != "/repos/a" {
		t.Fatalf("repo should come from config file, got %q", merged.RepoPath)
	}
	if merged.OutputFile != "out.md" {
		t.Fatalf("output should come from cli, got %q", merged.OutputFile)
	}
	if merged.Provider != ai.ProviderDeepSeek {
		t.Fatalf("provider should come from cli, got %q", merged.Provider)
	}
	if merged.BaseURL != ai.DefaultBaseURL(ai.ProviderDeepSeek) {
		t.Fatalf("base url should reset to the selected provider default, got %q", merged.BaseURL)
	}
	if merged.APIKey != "cli-key" {
		t.Fatalf("api key should come from cli, got %q", merged.APIKey)
	}
	if merged.Model != ai.DefaultModelName(ai.ProviderDeepSeek) {
		t.Fatalf("model should reset to the selected provider default, got %q", merged.Model)
	}
}

func TestMergeConfigResetsProviderDefaultsWhenProviderChanges(t *testing.T) {
	base := DefaultConfig()
	fileCfg := Config{
		Provider: ai.ProviderOpenAI,
		BaseURL:  ai.DefaultBaseURL(ai.ProviderOpenAI),
		Model:    ai.DefaultModelName(ai.ProviderOpenAI),
	}
	cli := Config{
		Provider: ai.ProviderGemini,
	}
	changed := map[string]bool{
		"provider": true,
	}

	merged := MergeConfig(base, fileCfg, cli, changed)
	if merged.Provider != ai.ProviderGemini {
		t.Fatalf("provider should be overridden by cli, got %q", merged.Provider)
	}
	if merged.BaseURL != ai.DefaultBaseURL(ai.ProviderGemini) {
		t.Fatalf("base url should reset to new provider default, got %q", merged.BaseURL)
	}
	if merged.Model != ai.DefaultModelName(ai.ProviderGemini) {
		t.Fatalf("model should reset to new provider default, got %q", merged.Model)
	}
}

func TestDefaultTimeSpec(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Time.Kind != timequery.KindPreset {
		t.Fatalf("unexpected default kind: %q", cfg.Time.Kind)
	}
	if cfg.Time.Period != timequery.PresetLast7Days {
		t.Fatalf("unexpected default period: %q", cfg.Time.Period)
	}
}

func TestNormalizeConfigPathsExpandsUserPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := Config{
		RepoPath:   "~/code/seakoi",
		OutputFile: "~/reports/out.md",
		Prompt:     "~/Documents/prompts/custom.md",
	}

	got, err := NormalizeConfigPaths(cfg)
	if err != nil {
		t.Fatalf("NormalizeConfigPaths returned error: %v", err)
	}

	if got.RepoPath != filepath.Join(home, "code", "seakoi") {
		t.Fatalf("repo path not expanded: %q", got.RepoPath)
	}
	if got.OutputFile != filepath.Join(home, "reports", "out.md") {
		t.Fatalf("output path not expanded: %q", got.OutputFile)
	}
	if got.Prompt != filepath.Join(home, "Documents", "prompts", "custom.md") {
		t.Fatalf("prompt path not expanded: %q", got.Prompt)
	}
}

func TestNormalizeConfigPathsKeepsBuiltInPrompt(t *testing.T) {
	cfg, err := NormalizeConfigPaths(Config{Prompt: "basic"})
	if err != nil {
		t.Fatalf("NormalizeConfigPaths returned error: %v", err)
	}
	if cfg.Prompt != "basic" {
		t.Fatalf("built-in prompt should stay unchanged, got %q", cfg.Prompt)
	}
}

func TestValidateConfigAllowsSparseConfig(t *testing.T) {
	if err := ValidateConfig(Config{}); err != nil {
		t.Fatalf("expected sparse config to be valid, got %v", err)
	}
}
