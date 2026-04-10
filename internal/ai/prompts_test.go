package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/celiumgrid/git-digest/internal/git"
	"github.com/celiumgrid/git-digest/internal/i18n"
)

func TestGetPromptType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PromptType
	}{
		{"基础提示词", "basic", BasicPrompt},
		{"向上汇报提示词", "manager-update", ManagerUpdatePrompt},
		{"自我复盘提示词", "self-review", SelfReviewPrompt},
		{"详细提示词", "detailed", DetailedPrompt},
		{"发布说明提示词", "release-notes", ReleaseNotesPrompt},
		{"未知类型", "unknown", PromptType("unknown")},
		{"空字符串", "", PromptType("")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetPromptTypeFromString(test.input)
			if result != test.expected {
				t.Errorf("输入 %s: 期望 %s, 得到 %s", test.input, test.expected, result)
			}
		})
	}
}

func TestLoadPromptTemplate(t *testing.T) {
	testContent := "测试模板内容 {{.CommitMessages}}"
	tmpfile, err := os.CreateTemp("", "test-prompt-*.txt")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatalf("无法写入临时文件: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("无法关闭临时文件: %v", err)
	}

	_, err = loadPromptTemplateFromPath("/non/existent/path.txt")
	if err == nil {
		t.Error("加载不存在的文件应该返回错误")
	}

	content, err := loadPromptTemplateFromPath(tmpfile.Name())
	if err != nil {
		t.Errorf("加载存在的文件不应该返回错误: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("加载的内容不匹配: 期望 %q, 得到 %q", testContent, string(content))
	}
}

func TestLoadCustomPromptExpandsTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, "Documents", "docs", "seakoi", "kpi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("无法创建目录: %v", err)
	}

	path := filepath.Join(dir, "kpi-prompt.md")
	if err := os.WriteFile(path, []byte("custom prompt"), 0o600); err != nil {
		t.Fatalf("无法写入提示词文件: %v", err)
	}

	content, err := LoadCustomPrompt("~/Documents/docs/seakoi/kpi/kpi-prompt.md", i18n.LanguageChinese)
	if err != nil {
		t.Fatalf("LoadCustomPrompt 返回错误: %v", err)
	}

	if content != "custom prompt\n" {
		t.Fatalf("unexpected prompt content: %q", content)
	}
}

func TestLoadBuiltInPromptDoesNotDependOnWorkingDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	})

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	content, err := loadPromptTemplate(BasicPrompt, i18n.LanguageEnglish)
	if err != nil {
		t.Fatalf("built-in prompt should still load outside the repo root: %v", err)
	}
	if content == "" {
		t.Fatal("built-in prompt should not be empty")
	}
}

func TestLoadPromptTemplateUsesLanguageSpecificBuiltins(t *testing.T) {
	english, err := loadPromptTemplate(ManagerUpdatePrompt, i18n.LanguageEnglish)
	if err != nil {
		t.Fatalf("failed to load english built-in prompt: %v", err)
	}

	chinese, err := loadPromptTemplate(ManagerUpdatePrompt, i18n.LanguageChinese)
	if err != nil {
		t.Fatalf("failed to load chinese built-in prompt: %v", err)
	}

	if english == chinese {
		t.Fatal("expected language-specific prompt templates to differ")
	}
	if !strings.Contains(english, "{{.CommitMessages}}") {
		t.Fatal("english prompt should keep commit placeholder")
	}
	if !strings.Contains(chinese, "{{.CommitMessages}}") {
		t.Fatal("chinese prompt should keep commit placeholder")
	}
}

func TestBuildPromptWithTemplateFailsForMissingCustomPrompt(t *testing.T) {
	_, err := buildPromptWithTemplate(
		[]git.CommitInfo{{
			Hash:    "12345678",
			Author:  "cola",
			Date:    time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC),
			Message: "feat: prompt test",
		}},
		time.Time{},
		time.Time{},
		PromptType("/tmp/does-not-exist.txt"),
		i18n.LanguageEnglish,
	)
	if err == nil {
		t.Fatal("expected missing custom prompt to return error")
	}
}
