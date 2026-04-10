package ai

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/celiumgrid/git-digest/internal/i18n"
	"github.com/celiumgrid/git-digest/internal/pathutil"
)

//go:embed prompts/*.txt
var builtInPromptFS embed.FS

// PromptType 表示不同类型的提示词
type PromptType string

const (
	// BasicPrompt 基础提示词：核心摘要
	BasicPrompt PromptType = "basic"
	// DetailedPrompt 中级提示词：详细且结构化的报告
	DetailedPrompt PromptType = "detailed"
	// TargetedPrompt 高级提示词：面向角色和受众的报告
	TargetedPrompt PromptType = "targeted"
)

// GetPromptTypeFromString 根据字符串返回对应的提示词类型
func GetPromptTypeFromString(promptTypeStr string) PromptType {
	switch promptTypeStr {
	case string(BasicPrompt):
		return BasicPrompt
	case string(DetailedPrompt):
		return DetailedPrompt
	case string(TargetedPrompt):
		return TargetedPrompt
	default:
		// 如果不是预设类型，返回作为自定义类型（文件路径）
		return PromptType(promptTypeStr)
	}
}

// IsCustomPrompt 检查是否为自定义提示词（文件路径）
func IsCustomPrompt(promptType PromptType) bool {
	return promptType != BasicPrompt &&
		promptType != DetailedPrompt &&
		promptType != TargetedPrompt
}

// LoadCustomPrompt 加载自定义提示词文件
func LoadCustomPrompt(filePath, language string) (string, error) {
	filePath, err := pathutil.NormalizeUserPath(filePath)
	if err != nil {
		if _, cwdErr := os.Getwd(); cwdErr != nil {
			return "", fmt.Errorf(i18n.T(language, "ai.getcwd"), cwdErr)
		}
		return "", fmt.Errorf(i18n.T(language, "ai.custom_prompt_read"), err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf(i18n.T(language, "ai.custom_prompt_missing"), filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf(i18n.T(language, "ai.custom_prompt_read"), err)
	}

	promptContent := strings.TrimSpace(string(content))
	if promptContent == "" {
		return "", fmt.Errorf(i18n.T(language, "ai.custom_prompt_empty"), filePath)
	}

	// 确保提示词末尾有一个换行符，以便后续添加提交记录
	if !strings.HasSuffix(promptContent, "\n") {
		promptContent += "\n"
	}

	return promptContent, nil
}

func loadBuiltInPrompt(filename string) (string, error) {
	content, err := builtInPromptFS.ReadFile(filepath.Join("prompts", filename))
	if err != nil {
		return "", err
	}
	return string(content), nil
}
