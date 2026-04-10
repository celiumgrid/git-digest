package app

import (
	"fmt"
	"testing"

	"github.com/celiumgrid/git-digest/internal/ai"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

type promptCall struct {
	kind  string
	label string
}

type stubPrompter struct {
	selects  []string
	inputs   []string
	confirms []bool
	calls    []promptCall
}

func (s *stubPrompter) SetLanguage(string) {}

func (s *stubPrompter) Select(label string, options []string, defaultValue string) (string, error) {
	s.calls = append(s.calls, promptCall{kind: "select", label: label})
	if len(s.selects) == 0 {
		return "", fmt.Errorf("unexpected select: %s", label)
	}
	value := s.selects[0]
	s.selects = s.selects[1:]
	return value, nil
}

func (s *stubPrompter) Input(label, defaultValue string, secret bool) (string, error) {
	s.calls = append(s.calls, promptCall{kind: "input", label: label})
	if len(s.inputs) == 0 {
		return defaultValue, nil
	}
	value := s.inputs[0]
	s.inputs = s.inputs[1:]
	return value, nil
}

func (s *stubPrompter) Confirm(label string, defaultValue bool) (bool, error) {
	s.calls = append(s.calls, promptCall{kind: "confirm", label: label})
	if len(s.confirms) == 0 {
		return defaultValue, nil
	}
	value := s.confirms[0]
	s.confirms = s.confirms[1:]
	return value, nil
}

func TestRunWizardUsesSelectForPresetFlow(t *testing.T) {
	prompter := &stubPrompter{
		selects:  []string{"English", "single", "preset", "this-month", "text", "basic", "Gemini", "gemini-2.5-pro"},
		inputs:   []string{".", "", "", "", ""},
		confirms: []bool{false},
	}

	cfg, err := runWizardWithPrompter(prompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != timequery.LanguageEnglish {
		t.Fatalf("unexpected language: %q", cfg.Language)
	}
	if cfg.Time.Kind != timequery.KindPreset || cfg.Time.Period != timequery.PresetThisMonth {
		t.Fatalf("unexpected time spec: %+v", cfg.Time)
	}
	if cfg.Prompt != "basic" {
		t.Fatalf("unexpected prompt: %q", cfg.Prompt)
	}
	if cfg.Provider != ai.ProviderGemini {
		t.Fatalf("unexpected provider: %q", cfg.Provider)
	}
	if cfg.Model != "gemini-2.5-pro" {
		t.Fatalf("unexpected model: %q", cfg.Model)
	}
	assertPromptKinds(t, prompter.calls,
		"select:Language（语言）",
		"select:Analysis mode",
		"input:Repository path (--repo)",
		"select:Time input",
		"select:Time period (--period)",
		"select:Output format",
		"select:Prompt type",
		"input:Author filter (--author, leave empty to use current Git user)",
		"input:Output file (--output, leave empty for stdout)",
		"select:AI provider",
		"input:Base URL (--base-url)",
		"input:API key (--api-key, leave empty to use environment variables)",
		"select:Model",
	)
}

func TestRunWizardUsesInputForSingleDayAndRange(t *testing.T) {
	dayPrompter := &stubPrompter{
		selects:  []string{"English", "single", "day", "text", "basic", "Gemini", "gemini-2.5-pro"},
		inputs:   []string{".", "2026-04-09", "", "", "", ""},
		confirms: []bool{false},
	}
	dayCfg, err := runWizardWithPrompter(dayPrompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dayCfg.Time.Kind != timequery.KindSingleDay || dayCfg.Time.On != "2026-04-09" {
		t.Fatalf("unexpected day time spec: %+v", dayCfg.Time)
	}

	rangePrompter := &stubPrompter{
		selects:  []string{"English", "single", "range", "text", "basic", "Gemini", "gemini-2.5-pro"},
		inputs:   []string{".", "2026-04-01", "2026-04-09", "", "", "", ""},
		confirms: []bool{false},
	}
	rangeCfg, err := runWizardWithPrompter(rangePrompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rangeCfg.Time.Kind != timequery.KindRange || rangeCfg.Time.From != "2026-04-01" || rangeCfg.Time.To != "2026-04-09" {
		t.Fatalf("unexpected range time spec: %+v", rangeCfg.Time)
	}
}

func TestRunWizardPromptsForCustomPromptPath(t *testing.T) {
	prompter := &stubPrompter{
		selects:  []string{"English", "single", "preset", "last-7d", "text", "custom-file", "OpenAI", "gpt-4.1-mini"},
		inputs:   []string{".", "/tmp/team.md", "", "", "", ""},
		confirms: []bool{false},
	}

	cfg, err := runWizardWithPrompter(prompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Prompt != "/tmp/team.md" {
		t.Fatalf("unexpected custom prompt path: %q", cfg.Prompt)
	}
	if cfg.Provider != ai.ProviderOpenAI {
		t.Fatalf("unexpected provider: %q", cfg.Provider)
	}
	assertContainsCall(t, prompter.calls, promptCall{kind: "input", label: "Custom prompt file path"})
}

func TestRunWizardPromptsForCustomModel(t *testing.T) {
	prompter := &stubPrompter{
		selects:  []string{"English", "single", "preset", "last-7d", "text", "basic", "DeepSeek", "custom-model"},
		inputs:   []string{".", "", "", "", "", "deepseek-reasoner"},
		confirms: []bool{false},
	}

	cfg, err := runWizardWithPrompter(prompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "deepseek-reasoner" {
		t.Fatalf("unexpected custom model: %q", cfg.Model)
	}
	assertContainsCall(t, prompter.calls, promptCall{kind: "input", label: "Custom model name"})
}

func TestRunWizardAcceptsNewBuiltInPromptTypes(t *testing.T) {
	prompter := &stubPrompter{
		selects:  []string{"English", "single", "preset", "last-7d", "text", "manager-update", "Gemini", "gemini-2.5-pro"},
		inputs:   []string{".", "", "", "", ""},
		confirms: []bool{false},
	}

	cfg, err := runWizardWithPrompter(prompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Prompt != "manager-update" {
		t.Fatalf("unexpected prompt: %q", cfg.Prompt)
	}
}

func TestRunWizardNoLongerPromptsToSaveBaseConfig(t *testing.T) {
	prompter := &stubPrompter{
		selects: []string{
			"English",
			"single",
			"preset",
			"this-month",
			"text",
			"basic",
			"Gemini",
			"gemini-2.5-pro",
		},
		inputs: []string{".", "", "", "", ""},
	}

	cfg, err := runWizardWithPrompter(prompter, DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prompter.confirms) != 0 {
		t.Fatalf("runtime wizard should not consume confirm prompts")
	}
	if cfg.Provider != ai.ProviderGemini {
		t.Fatalf("unexpected provider: %q", cfg.Provider)
	}
	for _, call := range prompter.calls {
		if call.kind == "confirm" {
			t.Fatalf("runtime wizard should not ask for save confirmation: %+v", call)
		}
	}
}

func TestRunBaseConfigWizardLeavesFieldsUnset(t *testing.T) {
	prompter := &stubPrompter{
		selects: []string{
			"Leave unset",
			"Leave unset",
			"Leave unset",
			"Leave unset",
			"Leave unset",
			"Leave unset",
		},
	}

	cfg, err := runBaseConfigWizardWithPrompter(prompter, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg != (Config{}) {
		t.Fatalf("expected empty sparse config, got %+v", cfg)
	}
}

func TestRunBaseConfigWizardBuildsSparseConfig(t *testing.T) {
	prompter := &stubPrompter{
		selects: []string{
			"中文",
			"多仓库",
			"预设周期",
			"上个月",
			"Markdown",
			"向上汇报",
			"OpenAI",
		},
		inputs: []string{
			"~/code",
			"alice",
			"~/reports/team.md",
			"https://proxy.example/v1",
			"sk-test",
			"gpt-4.1-mini",
		},
	}

	cfg, err := runBaseConfigWizardWithPrompter(prompter, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Language != timequery.LanguageChinese {
		t.Fatalf("unexpected language: %q", cfg.Language)
	}
	if cfg.RepoPath != "" || cfg.ReposPath != "~/code" {
		t.Fatalf("unexpected repo config: %+v", cfg)
	}
	if cfg.Time.Kind != timequery.KindPreset || cfg.Time.Period != timequery.PresetLastMonth {
		t.Fatalf("unexpected time config: %+v", cfg.Time)
	}
	if cfg.Format != "markdown" {
		t.Fatalf("unexpected format: %q", cfg.Format)
	}
	if cfg.Prompt != "manager-update" {
		t.Fatalf("unexpected prompt: %q", cfg.Prompt)
	}
	if cfg.Provider != ai.ProviderOpenAI {
		t.Fatalf("unexpected provider: %q", cfg.Provider)
	}
	if cfg.BaseURL != "https://proxy.example/v1" {
		t.Fatalf("unexpected base url: %q", cfg.BaseURL)
	}
	if cfg.APIKey != "sk-test" {
		t.Fatalf("unexpected api key: %q", cfg.APIKey)
	}
	if cfg.Model != "gpt-4.1-mini" {
		t.Fatalf("unexpected model: %q", cfg.Model)
	}
}

func assertPromptKinds(t *testing.T, calls []promptCall, want ...string) {
	t.Helper()
	if len(calls) != len(want) {
		t.Fatalf("call count mismatch: got %d want %d", len(calls), len(want))
	}
	for i, expected := range want {
		got := calls[i].kind + ":" + calls[i].label
		if got != expected {
			t.Fatalf("call %d mismatch: got %q want %q", i, got, expected)
		}
	}
}

func assertContainsCall(t *testing.T, calls []promptCall, want promptCall) {
	t.Helper()
	for _, call := range calls {
		if call == want {
			return
		}
	}
	t.Fatalf("missing call: %+v in %+v", want, calls)
}
