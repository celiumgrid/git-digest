package ai

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/celiumgrid/git-digest/internal/i18n"
)

func TestGetPromptType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PromptType
	}{
		{"基础提示词", "basic", BasicPrompt},
		{"详细提示词", "detailed", DetailedPrompt},
		{"针对性提示词", "targeted", TargetedPrompt},
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
