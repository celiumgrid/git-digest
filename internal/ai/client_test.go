package ai

import "testing"

func TestNormalizeClientConfigAppliesProviderDefaultsAndEnv(t *testing.T) {
	cfg, err := NormalizeClientConfig(ClientConfig{Provider: ProviderDeepSeek}, func(key string) string {
		if key == "DEEPSEEK_API_KEY" {
			return "deepseek-key"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != ProviderDeepSeek {
		t.Fatalf("provider mismatch: %q", cfg.Provider)
	}
	if cfg.BaseURL != DefaultBaseURL(ProviderDeepSeek) {
		t.Fatalf("base url mismatch: %q", cfg.BaseURL)
	}
	if cfg.Model != DefaultModelName(ProviderDeepSeek) {
		t.Fatalf("model mismatch: %q", cfg.Model)
	}
	if cfg.APIKey != "deepseek-key" {
		t.Fatalf("api key mismatch: %q", cfg.APIKey)
	}
}

func TestNormalizeClientConfigRejectsUnknownProvider(t *testing.T) {
	_, err := NormalizeClientConfig(ClientConfig{Provider: "unknown", APIKey: "x", Model: "m"}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNormalizeClientConfigAcceptsExplicitValues(t *testing.T) {
	cfg, err := NormalizeClientConfig(ClientConfig{
		Provider: ProviderOpenAI,
		BaseURL:  "https://proxy.example/v1",
		APIKey:   "explicit-key",
		Model:    "custom-model",
	}, func(string) string { return "" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://proxy.example/v1" {
		t.Fatalf("explicit base url lost: %q", cfg.BaseURL)
	}
	if cfg.APIKey != "explicit-key" {
		t.Fatalf("explicit api key lost: %q", cfg.APIKey)
	}
	if cfg.Model != "custom-model" {
		t.Fatalf("explicit model lost: %q", cfg.Model)
	}
}
