package app

import (
	"path/filepath"
	"testing"
)

func TestResolveRepoPathsExpandsTildeForSingleRepo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := resolveRepoPaths(Config{
		RepoPath: "~/code/seakoi",
	})
	if err != nil {
		t.Fatalf("resolveRepoPaths returned error: %v", err)
	}

	want := filepath.Join(home, "code", "seakoi")
	if len(got) != 1 {
		t.Fatalf("expected one repo path, got %d", len(got))
	}
	if got[0] != want {
		t.Fatalf("expected expanded repo path %q, got %q", want, got[0])
	}
}
