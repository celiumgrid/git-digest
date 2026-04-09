package i18n

import (
	"strings"
	"testing"
)

func TestBaseConfigTerminology(t *testing.T) {
	englishKeys := []string{"main.saved_config", "flag.no_config", "flag.save_config", "wizard.save_default"}
	for _, key := range englishKeys {
		if !strings.Contains(strings.ToLower(T(LanguageEnglish, key)), "base config") &&
			!strings.Contains(strings.ToLower(T(LanguageEnglish, key)), "base configuration") {
			t.Fatalf("expected english %s to mention base config, got %q", key, T(LanguageEnglish, key))
		}
	}

	chineseKeys := []string{"main.saved_config", "flag.no_config", "flag.save_config", "wizard.save_default"}
	for _, key := range chineseKeys {
		if !strings.Contains(T(LanguageChinese, key), "基础配置") {
			t.Fatalf("expected chinese %s to mention 基础配置, got %q", key, T(LanguageChinese, key))
		}
	}
}
