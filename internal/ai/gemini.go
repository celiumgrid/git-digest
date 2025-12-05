package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/kway-teow/git-work-log/internal/git"
	"google.golang.org/api/option"
)

// 默认模型名称
const DefaultModelName = "gemini-2.5-pro"

// GeminiClient 是Gemini AI API的客户端
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiClient 创建一个新的Gemini客户端
func NewGeminiClient() (*GeminiClient, error) {
	return NewGeminiClientWithModel(DefaultModelName) // 默认使用gemini-2.5-pro模型
}

// NewGeminiClientWithModel 使用指定模型创建一个新的Gemini客户端
func NewGeminiClientWithModel(modelName string) (*GeminiClient, error) {
	// 如果没有指定模型名称，使用默认模型
	if modelName == "" {
		modelName = DefaultModelName
	}
	// 从环境变量获取API密钥
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("未设置GEMINI_API_KEY环境变量")
	}

	// 创建Gemini客户端
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("创建Gemini客户端失败: %w", err)
	}

	// 使用指定的模型
	model := client.GenerativeModel(modelName)

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// SummarizeCommits 使用AI总结提交记录
func (g *GeminiClient) SummarizeCommits(commits []git.CommitInfo) (string, error) {
	return g.SummarizeCommitsWithPrompt(commits, BasicPrompt)
}

// SummarizeCommitsWithPrompt 使用指定的提示词类型总结提交记录
func (g *GeminiClient) SummarizeCommitsWithPrompt(commits []git.CommitInfo, promptType PromptType) (string, error) {
	if len(commits) == 0 {
		return "没有找到提交记录。", nil
	}

	// 获取时间范围
	var earliestDate, latestDate time.Time
	if len(commits) > 0 {
		earliestDate = commits[len(commits)-1].Date
		latestDate = commits[0].Date

		// 遍历所有提交，找出最早和最晚的日期
		for _, commit := range commits {
			if commit.Date.Before(earliestDate) {
				earliestDate = commit.Date
			}
			if commit.Date.After(latestDate) {
				latestDate = commit.Date
			}
		}
	}

	// 构建提示词
	prompt := buildPromptWithTemplate(commits, earliestDate, latestDate, promptType)

	// 调用Gemini API
	ctx := context.Background()
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("调用Gemini API失败: %w", err)
	}

	// 提取回复
	var result strings.Builder
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			result.WriteString(fmt.Sprintf("%v", part))
		}
	}

	return result.String(), nil
}

// buildPromptWithTemplate 使用指定的提示词模板构建提示词
func buildPromptWithTemplate(commits []git.CommitInfo, _ /*fromDate*/, _ /*toDate*/ time.Time, promptType PromptType) string {
	// 获取提示词模板
	template, err := loadPromptTemplate(promptType)
	if err != nil {
		// 如果加载模板失败，使用默认的提示词
		fmt.Printf("警告: 加载提示词模板失败: %v, 使用默认提示词\n", err)
		template = defaultPromptTemplate
	}

	// 构建提交记录字符串
	var commitMessages strings.Builder

	for i, commit := range commits {
		// 添加提交记录
		fmt.Fprintf(&commitMessages, "提交 %d:\n", i+1)
		fmt.Fprintf(&commitMessages, "- 哈希值: %s\n", commit.Hash[:8])
		fmt.Fprintf(&commitMessages, "- 作者: %s\n", commit.Author)
		fmt.Fprintf(&commitMessages, "- 日期: %s\n", commit.Date.Format("2006-01-02 15:04:05"))

		// 添加分支信息
		if len(commit.Branches) > 0 {
			fmt.Fprintf(&commitMessages, "- 分支: %s\n", strings.Join(commit.Branches, ", "))
		}

		// 添加提交消息
		fmt.Fprintf(&commitMessages, "- 消息: %s\n", commit.Message)

		// 添加变更文件
		if len(commit.ChangedFiles) > 0 {
			fmt.Fprintf(&commitMessages, "- 变更文件:\n")
			// 最多显示10个文件
			maxFiles := 10
			if len(commit.ChangedFiles) < maxFiles {
				maxFiles = len(commit.ChangedFiles)
			}
			for j := 0; j < maxFiles; j++ {
				fmt.Fprintf(&commitMessages, "  * %s\n", commit.ChangedFiles[j])
			}
			if len(commit.ChangedFiles) > maxFiles {
				fmt.Fprintf(&commitMessages, "  * ... 以及其他 %d 个文件\n", len(commit.ChangedFiles)-maxFiles)
			}
		}

		// 添加空行分隔不同提交
		fmt.Fprintf(&commitMessages, "\n")
	}

	// 替换模板中的变量
	prompt := strings.ReplaceAll(template, "{{.CommitMessages}}", commitMessages.String())

	return prompt
}

// GenerateReport 根据提交记录和时间范围生成报告
func (g *GeminiClient) GenerateReport(commits []git.CommitInfo, fromDate, toDate time.Time) (string, error) {
	return g.GenerateReportWithPrompt(commits, fromDate, toDate, BasicPrompt)
}

// GenerateReportWithPrompt 使用指定的提示词类型生成报告
func (g *GeminiClient) GenerateReportWithPrompt(commits []git.CommitInfo, fromDate, toDate time.Time, promptType PromptType) (string, error) {
	// 这个方法实际上是对SummarizeCommits的封装，提供更明确的接口
	if len(commits) == 0 {
		// 根据时间范围返回不同的消息
		daysDiff := toDate.Sub(fromDate).Hours() / 24
		var periodType string

		switch {
		case daysDiff <= 1:
			periodType = "今日"
		case daysDiff <= 7:
			periodType = "本周"
		case daysDiff <= 31:
			periodType = "本月"
		case daysDiff <= 366:
			periodType = "本年"
		default:
			periodType = "指定时间范围内"
		}

		return fmt.Sprintf("%s没有提交记录。", periodType), nil
	}

	// 构建提示词
	prompt := buildPromptWithTemplate(commits, fromDate, toDate, promptType)

	// 调用Gemini API
	ctx := context.Background()
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("调用Gemini API失败: %w", err)
	}

	// 提取回复
	var result strings.Builder
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			result.WriteString(fmt.Sprintf("%v", part))
		}
	}

	return result.String(), nil
}

// Close 关闭Gemini客户端
func (g *GeminiClient) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

// 默认提示词模板
const defaultPromptTemplate = `你是一位专业的工作报告生成助手。请根据以下Git提交记录，生成一份简洁的工作摘要。

提交记录：
{{.CommitMessages}}

请提供：
1. 一个简短的总体工作概述（不超过3句话）
2. 3-5个关键工作成就或完成的任务
3. 任何明显的工作主题或模式

保持简洁明了，重点突出实际完成的工作。`

// loadPromptTemplate 从文件加载提示词模板
func loadPromptTemplate(promptType PromptType) (string, error) {
	// 检查是否为自定义提示词
	if IsCustomPrompt(promptType) {
		// 自定义提示词，从文件加载
		customPrompt, err := LoadCustomPrompt(string(promptType))
		if err != nil {
			return "", fmt.Errorf("加载自定义提示词失败: %w", err)
		}

		// 确保自定义提示词包含占位符
		if !strings.Contains(customPrompt, "{{.CommitMessages}}") {
			// 如果没有占位符，在末尾添加
			customPrompt += "\n\n提交记录：\n{{.CommitMessages}}"
		}

		return customPrompt, nil
	}

	// 预设提示词，根据提示词类型确定文件名
	var filename string
	switch promptType {
	case BasicPrompt:
		filename = "basic.txt"
	case DetailedPrompt:
		filename = "detailed.txt"
	case TargetedPrompt:
		filename = "targeted.txt"
	default:
		filename = "basic.txt"
	}

	// 尝试从多个可能的位置加载模板
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}

	// 尝试从多个可能的位置加载模板
	paths := []string{
		fmt.Sprintf("%s/prompts/%s", cwd, filename),                                 // 当前目录下的prompts目录
		fmt.Sprintf("%s/../prompts/%s", cwd, filename),                              // 上级目录下的prompts目录
		fmt.Sprintf("%s/../../prompts/%s", cwd, filename),                           // 上上级目录下的prompts目录
		fmt.Sprintf("/Users/cola/code/kway-teow/git-work-log/prompts/%s", filename), // 项目根目录
	}

	var content []byte
	var loadErr error

	// 尝试每个路径
	for _, path := range paths {
		content, loadErr = loadPromptTemplateFromPath(path)
		if loadErr == nil {
			// 成功加载模板
			return string(content), nil
		}
	}

	// 所有路径都失败了，返回最后一个错误
	return "", fmt.Errorf("无法加载提示词模板: %w", loadErr)
}

// loadPromptTemplateFromPath 从指定路径加载提示词模板
func loadPromptTemplateFromPath(path string) ([]byte, error) {
	return os.ReadFile(path)
}
