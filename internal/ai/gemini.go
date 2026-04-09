package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kway-teow/git-digest/internal/git"
	"github.com/kway-teow/git-digest/internal/i18n"
	openai "github.com/sashabaranov/go-openai"
)

const (
	ProviderOpenAI   = "openai"
	ProviderGemini   = "gemini"
	ProviderDeepSeek = "deepseek"

	DefaultProvider = ProviderGemini
)

// ClientConfig is the provider-neutral runtime config for the AI client.
type ClientConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
	Language string
}

// DefaultModelName returns the default text model for a provider.
func DefaultModelName(provider string) string {
	switch normalizeProvider(provider) {
	case ProviderOpenAI:
		return "gpt-4.1-mini"
	case ProviderDeepSeek:
		return "deepseek-chat"
	case ProviderGemini:
		fallthrough
	default:
		return "gemini-2.5-pro"
	}
}

// DefaultBaseURL returns the default OpenAI-compatible endpoint for a provider.
func DefaultBaseURL(provider string) string {
	switch normalizeProvider(provider) {
	case ProviderOpenAI:
		return "https://api.openai.com/v1"
	case ProviderDeepSeek:
		return "https://api.deepseek.com/v1"
	case ProviderGemini:
		fallthrough
	default:
		return "https://generativelanguage.googleapis.com/v1beta/openai"
	}
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return DefaultProvider
	}
	return provider
}

func lookupAPIKey(provider string, lookup func(string) string) string {
	for _, key := range envKeysForProvider(provider) {
		if value := strings.TrimSpace(lookup(key)); value != "" {
			return value
		}
	}
	return ""
}

func envKeysForProvider(provider string) []string {
	switch normalizeProvider(provider) {
	case ProviderOpenAI:
		return []string{"OPENAI_API_KEY"}
	case ProviderDeepSeek:
		return []string{"DEEPSEEK_API_KEY", "OPENAI_API_KEY"}
	case ProviderGemini:
		return []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"}
	default:
		return nil
	}
}

// NormalizeClientConfig applies provider defaults and resolves env fallback values.
func NormalizeClientConfig(cfg ClientConfig, lookup func(string) string) (ClientConfig, error) {
	provider := normalizeProvider(cfg.Provider)
	if lookup == nil {
		lookup = os.Getenv
	}

	switch provider {
	case ProviderOpenAI, ProviderGemini, ProviderDeepSeek:
	default:
		return ClientConfig{}, fmt.Errorf(i18n.T(cfg.Language, "ai.unsupported_provider"), cfg.Provider)
	}

	cfg.Provider = provider
	cfg.Language = i18n.NormalizeLanguage(cfg.Language)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL(provider)
	}
	cfg.Model = strings.TrimSpace(cfg.Model)
	if cfg.Model == "" {
		cfg.Model = DefaultModelName(provider)
	}
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	if cfg.APIKey == "" {
		cfg.APIKey = lookupAPIKey(provider, lookup)
	}
	if cfg.APIKey == "" {
		return ClientConfig{}, fmt.Errorf(i18n.T(cfg.Language, "ai.api_key_missing"), strings.Join(envKeysForProvider(provider), "/"))
	}

	return cfg, nil
}

// Client wraps an OpenAI-compatible chat completion client.
type Client struct {
	client *openai.Client
	cfg    ClientConfig
}

func NewClient(cfg ClientConfig) (*Client, error) {
	normalized, err := NormalizeClientConfig(cfg, os.Getenv)
	if err != nil {
		return nil, err
	}

	oaiCfg := openai.DefaultConfig(normalized.APIKey)
	oaiCfg.BaseURL = normalized.BaseURL

	return &Client{
		client: openai.NewClientWithConfig(oaiCfg),
		cfg:    normalized,
	}, nil
}

func NewGeminiClient() (*Client, error) {
	return NewClient(ClientConfig{Provider: ProviderGemini, Language: i18n.LanguageEnglish})
}

func NewGeminiClientWithModel(modelName string) (*Client, error) {
	return NewClient(ClientConfig{Provider: ProviderGemini, Model: modelName, Language: i18n.LanguageEnglish})
}

func (c *Client) SummarizeCommits(commits []git.CommitInfo) (string, error) {
	return c.SummarizeCommitsWithPrompt(commits, BasicPrompt)
}

func (c *Client) SummarizeCommitsWithPrompt(commits []git.CommitInfo, promptType PromptType) (string, error) {
	if len(commits) == 0 {
		return i18n.T(c.cfg.Language, "ai.no_commits"), nil
	}

	var earliestDate, latestDate time.Time
	earliestDate = commits[len(commits)-1].Date
	latestDate = commits[0].Date
	for _, commit := range commits {
		if commit.Date.Before(earliestDate) {
			earliestDate = commit.Date
		}
		if commit.Date.After(latestDate) {
			latestDate = commit.Date
		}
	}

	prompt := buildPromptWithTemplate(commits, earliestDate, latestDate, promptType, c.cfg.Language)
	return c.generate(prompt)
}

func (c *Client) GenerateReport(commits []git.CommitInfo, fromDate, toDate time.Time) (string, error) {
	return c.GenerateReportWithPrompt(commits, fromDate, toDate, BasicPrompt)
}

func (c *Client) GenerateReportWithPrompt(commits []git.CommitInfo, fromDate, toDate time.Time, promptType PromptType) (string, error) {
	if len(commits) == 0 {
		daysDiff := toDate.Sub(fromDate).Hours() / 24

		switch {
		case daysDiff <= 1:
			return i18n.T(c.cfg.Language, "ai.no_commits_day"), nil
		case daysDiff <= 7:
			return i18n.T(c.cfg.Language, "ai.no_commits_week"), nil
		case daysDiff <= 31:
			return i18n.T(c.cfg.Language, "ai.no_commits_month"), nil
		case daysDiff <= 366:
			return i18n.T(c.cfg.Language, "ai.no_commits_year"), nil
		default:
			return i18n.T(c.cfg.Language, "ai.no_commits_range"), nil
		}
	}

	prompt := buildPromptWithTemplate(commits, fromDate, toDate, promptType, c.cfg.Language)
	return c.generate(prompt)
}

func (c *Client) generate(prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: c.cfg.Model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf(i18n.T(c.cfg.Language, "ai.chat_failed"), c.cfg.Provider, err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(i18n.T(c.cfg.Language, "ai.empty_response"), c.cfg.Provider)
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (c *Client) Close() {}

func buildPromptWithTemplate(commits []git.CommitInfo, _ /*fromDate*/, _ /*toDate*/ time.Time, promptType PromptType, language string) string {
	template, err := loadPromptTemplate(promptType, language)
	if err != nil {
		fmt.Printf(i18n.T(language, "ai.prompt_load_warning")+"\n", err)
		template = i18n.T(language, "ai.prompt_template")
	}

	var commitMessages strings.Builder
	for i, commit := range commits {
		fmt.Fprintf(&commitMessages, i18n.T(language, "ai.commit")+"\n", i+1)
		fmt.Fprintf(&commitMessages, i18n.T(language, "ai.hash")+"\n", commit.Hash[:8])
		fmt.Fprintf(&commitMessages, i18n.T(language, "ai.author")+"\n", commit.Author)
		fmt.Fprintf(&commitMessages, i18n.T(language, "ai.date")+"\n", commit.Date.Format("2006-01-02 15:04:05"))
		if len(commit.Branches) > 0 {
			fmt.Fprintf(&commitMessages, i18n.T(language, "ai.branch")+"\n", strings.Join(commit.Branches, ", "))
		}
		fmt.Fprintf(&commitMessages, i18n.T(language, "ai.message")+"\n", commit.Message)
		if len(commit.ChangedFiles) > 0 {
			fmt.Fprintln(&commitMessages, i18n.T(language, "ai.changed_files"))
			maxFiles := 10
			if len(commit.ChangedFiles) < maxFiles {
				maxFiles = len(commit.ChangedFiles)
			}
			for j := 0; j < maxFiles; j++ {
				fmt.Fprintf(&commitMessages, "  * %s\n", commit.ChangedFiles[j])
			}
			if len(commit.ChangedFiles) > maxFiles {
				fmt.Fprintf(&commitMessages, i18n.T(language, "ai.and_more_files")+"\n", len(commit.ChangedFiles)-maxFiles)
			}
		}
		fmt.Fprintf(&commitMessages, "\n")
	}

	return strings.ReplaceAll(template, "{{.CommitMessages}}", commitMessages.String())
}

func loadPromptTemplate(promptType PromptType, language string) (string, error) {
	if IsCustomPrompt(promptType) {
		customPrompt, err := LoadCustomPrompt(string(promptType), language)
		if err != nil {
			return "", fmt.Errorf(i18n.T(language, "ai.load_custom_prompt"), err)
		}
		if !strings.Contains(customPrompt, "{{.CommitMessages}}") {
			customPrompt += "\n\n" + i18n.T(language, "ai.commit_history") + ":\n{{.CommitMessages}}"
		}
		return customPrompt, nil
	}

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

	paths := candidatePromptPaths(filename)
	var content []byte
	var loadErr error
	for _, path := range paths {
		content, loadErr = loadPromptTemplateFromPath(path)
		if loadErr == nil {
			return string(content), nil
		}
	}

	return "", fmt.Errorf(i18n.T(language, "ai.load_custom_prompt"), loadErr)
}

func loadPromptTemplateFromPath(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func candidatePromptPaths(filename string) []string {
	roots := make([]string, 0, 8)
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, cwd, filepath.Dir(cwd), filepath.Dir(filepath.Dir(cwd)))
	}
	if execPath, err := os.Executable(); err == nil {
		binDir := filepath.Dir(execPath)
		roots = append(roots, binDir, filepath.Dir(binDir), filepath.Dir(filepath.Dir(binDir)))
	}

	paths := make([]string, 0, len(roots))
	seen := make(map[string]struct{})
	for _, root := range roots {
		if root == "" {
			continue
		}
		p := filepath.Join(root, "prompts", filename)
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}
	return paths
}
