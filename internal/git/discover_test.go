package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/celiumgrid/git-digest/internal/i18n"
)

func TestDiscoverGitReposExpandsTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoPath := filepath.Join(home, "code", "seakoi")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create test repo: %v", err)
	}

	repos, err := DiscoverGitRepos("~/code/seakoi", i18n.LanguageEnglish)
	if err != nil {
		t.Fatalf("DiscoverGitRepos returned error: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected one repo, got %d", len(repos))
	}
	if repos[0] != repoPath {
		t.Fatalf("expected expanded repo path %q, got %q", repoPath, repos[0])
	}
}
