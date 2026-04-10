package report

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/celiumgrid/git-digest/internal/git"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

type Format string

const (
	FormatText     Format = "text"
	FormatMarkdown Format = "markdown"
)

type Generator struct {
	Format Format
	Output io.Writer
}

func NewGenerator(format Format, output io.Writer) *Generator {
	if output == nil {
		output = os.Stdout
	}
	return &Generator{Format: format, Output: output}
}

func (g *Generator) GenerateReport(summary string, commits []git.CommitInfo, window timequery.Window, language string) error {
	language = timequery.NormalizeLanguage(language)
	switch g.Format {
	case FormatMarkdown:
		return g.generateMarkdownReport(summary, commits, window, language)
	default:
		return g.generateTextReport(summary, commits, window, language)
	}
}

func (g *Generator) generateTextReport(summary string, commits []git.CommitInfo, window timequery.Window, language string) error {
	fmt.Fprintf(g.Output, "%s (%s %s %s)\n", window.Label, window.Start.Format("2006-01-02"), word(language, "to", "至"), window.End.Format("2006-01-02"))
	fmt.Fprintln(g.Output, "==================================")
	fmt.Fprintln(g.Output)

	repoStats := make(map[string]int)
	for _, commit := range commits {
		if commit.RepoPath != "" {
			repoStats[commit.RepoPath]++
		}
	}

	if len(repoStats) > 1 {
		fmt.Fprintln(g.Output, heading(language, "Repo Stats", "仓库统计", 2))
		for repo, count := range repoStats {
			fmt.Fprintf(g.Output, "- %s: %d %s\n", repo, count, word(language, "commits", "条提交"))
		}
		fmt.Fprintln(g.Output)
	}

	fmt.Fprintln(g.Output, heading(language, "AI Summary", "AI 总结", 2))
	fmt.Fprintln(g.Output, summary)
	fmt.Fprintln(g.Output)
	fmt.Fprintln(g.Output, heading(language, "Commit History", "提交记录", 2))
	fmt.Fprintf(g.Output, "%s %d %s\n\n", word(language, "Total", "共有"), len(commits), word(language, "commits", "条提交记录"))

	for i, commit := range commits {
		fmt.Fprintf(g.Output, "%s %d:\n", word(language, "Commit", "提交"), i+1)
		fmt.Fprintf(g.Output, "- %s: %s\n", word(language, "Hash", "哈希值"), commit.Hash[:8])
		fmt.Fprintf(g.Output, "- %s: %s\n", word(language, "Author", "作者"), commit.Author)
		fmt.Fprintf(g.Output, "- %s: %s\n", word(language, "Date", "日期"), commit.Date.Format("2006-01-02 15:04:05"))
		if len(repoStats) > 1 && commit.RepoPath != "" {
			fmt.Fprintf(g.Output, "- %s: %s\n", word(language, "Repo", "仓库"), commit.RepoPath)
		}
		if len(commit.Branches) > 0 {
			fmt.Fprintf(g.Output, "- %s: %s\n", word(language, "Branch", "分支"), strings.Join(commit.Branches, ", "))
		}
		fmt.Fprintf(g.Output, "- %s: %s\n\n", word(language, "Message", "消息"), commit.Message)
	}

	return nil
}

func (g *Generator) generateMarkdownReport(summary string, commits []git.CommitInfo, window timequery.Window, language string) error {
	fmt.Fprintf(g.Output, "# %s (%s %s %s)\n\n", window.Label, window.Start.Format("2006-01-02"), word(language, "to", "至"), window.End.Format("2006-01-02"))

	repoStats := make(map[string]int)
	for _, commit := range commits {
		if commit.RepoPath != "" {
			repoStats[commit.RepoPath]++
		}
	}

	if len(repoStats) > 1 {
		fmt.Fprintln(g.Output, heading(language, "Repo Stats", "仓库统计", 2))
		fmt.Fprintln(g.Output)
		for repo, count := range repoStats {
			fmt.Fprintf(g.Output, "- **%s**: %d %s\n", repo, count, word(language, "commits", "条提交"))
		}
		fmt.Fprintln(g.Output)
	}

	fmt.Fprintln(g.Output, heading(language, "AI Summary", "AI 总结", 2))
	fmt.Fprintln(g.Output, summary)
	fmt.Fprintln(g.Output)
	fmt.Fprintln(g.Output, heading(language, "Commit History", "提交记录", 2))
	fmt.Fprintf(g.Output, "%s %d %s\n\n", word(language, "Total", "共有"), len(commits), word(language, "commits", "条提交记录"))

	for i, commit := range commits {
		fmt.Fprintf(g.Output, "%s %d\n\n", heading(language, "Commit", "提交", 3), i+1)
		fmt.Fprintf(g.Output, "- **%s**: `%s`\n", word(language, "Hash", "哈希值"), commit.Hash[:8])
		fmt.Fprintf(g.Output, "- **%s**: %s\n", word(language, "Author", "作者"), commit.Author)
		fmt.Fprintf(g.Output, "- **%s**: %s\n", word(language, "Date", "日期"), commit.Date.Format("2006-01-02 15:04:05"))
		if len(repoStats) > 1 && commit.RepoPath != "" {
			fmt.Fprintf(g.Output, "- **%s**: `%s`\n", word(language, "Repo", "仓库"), commit.RepoPath)
		}
		if len(commit.Branches) > 0 {
			fmt.Fprintf(g.Output, "- **%s**: %s\n", word(language, "Branch", "分支"), strings.Join(commit.Branches, ", "))
		}
		fmt.Fprintf(g.Output, "- **%s**: %s\n", word(language, "Message", "消息"), commit.Message)
		if len(commit.ChangedFiles) > 0 {
			fmt.Fprintln(g.Output, "- **"+word(language, "Changed Files", "变更文件")+"**:")
			for _, fileName := range commit.ChangedFiles {
				fmt.Fprintf(g.Output, "  - `%s`\n", fileName)
			}
		}
		fmt.Fprintln(g.Output)
	}

	return nil
}

func word(language, en, zh string) string {
	if timequery.NormalizeLanguage(language) == timequery.LanguageChinese {
		return zh
	}
	return en
}

func heading(language, en, zh string, level int) string {
	prefix := strings.Repeat("#", level)
	return prefix + " " + word(language, en, zh)
}
