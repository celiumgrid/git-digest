package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/celiumgrid/git-digest/internal/ai"
	"github.com/celiumgrid/git-digest/internal/git"
	"github.com/celiumgrid/git-digest/internal/i18n"
	"github.com/celiumgrid/git-digest/internal/report"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

type Service struct {
	Stdout io.Writer
	Stderr io.Writer
}

func NewService(stdout, stderr io.Writer) *Service {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &Service{Stdout: stdout, Stderr: stderr}
}

func (s *Service) Run(cfg Config) error {
	cfg.Language = i18n.NormalizeLanguage(cfg.Language)
	if err := ValidateConfig(cfg); err != nil {
		return err
	}

	window, err := timequery.ResolveWithLanguage(cfg.Time, time.Local, time.Now(), cfg.Language)
	if err != nil {
		return err
	}

	repoPaths, err := resolveRepoPaths(cfg)
	if err != nil {
		return err
	}

	fmt.Fprintf(s.Stdout, i18n.T(cfg.Language, "service.processing")+"\n", len(repoPaths))

	var allCommits []git.CommitInfo
	repoCommitCounts := make(map[string]int)

	for _, currentRepoPath := range repoPaths {
		fmt.Fprintf(s.Stdout, i18n.T(cfg.Language, "service.analyzing_repo")+"\n", currentRepoPath)
		gitOpts := git.NewGitOptionsWithLanguage(currentRepoPath, cfg.Language)
		if cfg.Author != "" {
			gitOpts.Author = cfg.Author
		}

		commits, commitErr := git.GetCommitsBetween(window.Start, window.End, gitOpts)
		if commitErr != nil {
			fmt.Fprintf(s.Stderr, i18n.T(cfg.Language, "service.warning_repo")+"\n", currentRepoPath, commitErr)
			continue
		}

		repoCommitCounts[currentRepoPath] = len(commits)
		for i := range commits {
			commits[i].RepoPath = currentRepoPath
		}
		allCommits = append(allCommits, commits...)
	}

	sort.Slice(allCommits, func(i, j int) bool {
		return allCommits[i].Date.After(allCommits[j].Date)
	})

	printStats(s.Stdout, cfg, repoCommitCounts)

	if len(allCommits) == 0 {
		fmt.Fprintf(s.Stdout, i18n.T(cfg.Language, "service.no_commits")+"\n", window.Start.Format("2006-01-02"), window.End.Format("2006-01-02"))
		return nil
	}

	client, err := ai.NewClient(ai.ClientConfig{
		Provider: cfg.Provider,
		BaseURL:  cfg.BaseURL,
		APIKey:   cfg.APIKey,
		Model:    cfg.Model,
		Language: cfg.Language,
	})
	if err != nil {
		return fmt.Errorf(i18n.T(cfg.Language, "service.create_ai_client"), err)
	}
	defer client.Close()

	aiPromptType := ai.GetPromptTypeFromString(cfg.Prompt)
	summary, err := client.SummarizeCommitsWithPrompt(allCommits, aiPromptType)
	if err != nil {
		return fmt.Errorf(i18n.T(cfg.Language, "service.generate_summary"), err)
	}

	var output = s.Stdout
	if cfg.OutputFile != "" {
		f, err := os.Create(cfg.OutputFile)
		if err != nil {
			return fmt.Errorf(i18n.T(cfg.Language, "service.create_output"), err)
		}
		defer f.Close()
		output = f
	}

	rg := report.NewGenerator(report.Format(cfg.Format), output)
	if err := rg.GenerateReport(summary, allCommits, window, cfg.Language); err != nil {
		return fmt.Errorf(i18n.T(cfg.Language, "service.write_report"), err)
	}

	fmt.Fprintln(s.Stdout, i18n.T(cfg.Language, "service.report_generated"))
	return nil
}

func resolveRepoPaths(cfg Config) ([]string, error) {
	switch {
	case cfg.ReposPath != "":
		repos, err := git.DiscoverGitRepos(cfg.ReposPath, cfg.Language)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(cfg.Language, "service.discover_repos"), err)
		}
		if len(repos) == 0 {
			return nil, fmt.Errorf(i18n.T(cfg.Language, "service.no_repos"), cfg.ReposPath)
		}
		return repos, nil
	case cfg.RepoPath != "":
		repoPath, err := git.NormalizePath(cfg.RepoPath)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(cfg.Language, "git.abs_path"), err)
		}
		return []string{repoPath}, nil
	default:
		return []string{"."}, nil
	}
}

func printStats(out io.Writer, cfg Config, repoCommitCounts map[string]int) {
	fmt.Fprintln(out, i18n.T(cfg.Language, "service.stats_heading"))
	total := 0
	for repoPath, count := range repoCommitCounts {
		displayPath := repoPath
		if cfg.ReposPath != "" {
			if rel, err := filepath.Rel(cfg.ReposPath, repoPath); err == nil {
				displayPath = rel
			}
		}
		fmt.Fprintf(out, i18n.T(cfg.Language, "service.stats_item")+"\n", displayPath, count)
		total += count
	}
	fmt.Fprintf(out, i18n.T(cfg.Language, "service.stats_total"), total)
	fmt.Fprintln(out)
}
