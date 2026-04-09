package app

import (
	"encoding/json"
	"os"
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

func TestLoadAndSaveConfigExpandTildePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgPath := "~/config/git-digest/config.json"
	wantPath := filepath.Join(home, "config", "git-digest", "config.json")

	cfg := Config{Prompt: "basic", Format: "text"}
	if err := SaveConfig(cfgPath, cfg, "en"); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}

	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected config file at %q: %v", wantPath, err)
	}

	raw, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	var stored Config
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("failed to parse saved config: %v", err)
	}
	if stored.Prompt != "basic" {
		t.Fatalf("unexpected stored prompt: %q", stored.Prompt)
	}

	loaded, err := LoadConfig(cfgPath, "en")
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if loaded.Prompt != "basic" {
		t.Fatalf("unexpected loaded prompt: %q", loaded.Prompt)
	}
}
