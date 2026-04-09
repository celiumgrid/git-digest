package main

import (
	"bytes"
	"runtime/debug"
	"testing"
)

func TestResolvedBuildMetadataFallsBackToBuildInfo(t *testing.T) {
	originalVersion, originalCommit, originalDate := version, commit, date
	originalReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		version, commit, date = originalVersion, originalCommit, originalDate
		readBuildInfo = originalReadBuildInfo
	})

	version = "dev"
	commit = "none"
	date = "unknown"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v1.2.3"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.time", Value: "2026-04-09T12:00:00Z"},
			},
		}, true
	}

	gotVersion, gotCommit, gotDate := resolvedBuildMetadata()
	if gotVersion != "v1.2.3" {
		t.Fatalf("expected build-info version fallback, got %q", gotVersion)
	}
	if gotCommit != "abc123" {
		t.Fatalf("expected build-info commit fallback, got %q", gotCommit)
	}
	if gotDate != "2026-04-09T12:00:00Z" {
		t.Fatalf("expected build-info date fallback, got %q", gotDate)
	}
}

func TestResolvedBuildMetadataPrefersInjectedValues(t *testing.T) {
	originalVersion, originalCommit, originalDate := version, commit, date
	originalReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		version, commit, date = originalVersion, originalCommit, originalDate
		readBuildInfo = originalReadBuildInfo
	})

	version = "v9.9.9"
	commit = "deadbeef"
	date = "2026-04-09T00:00:00Z"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v1.2.3"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.time", Value: "2026-04-09T12:00:00Z"},
			},
		}, true
	}

	gotVersion, gotCommit, gotDate := resolvedBuildMetadata()
	if gotVersion != "v9.9.9" {
		t.Fatalf("expected injected version to win, got %q", gotVersion)
	}
	if gotCommit != "deadbeef" {
		t.Fatalf("expected injected commit to win, got %q", gotCommit)
	}
	if gotDate != "2026-04-09T00:00:00Z" {
		t.Fatalf("expected injected date to win, got %q", gotDate)
	}
}

func TestWriteVersionOutputPrintsOnlyResolvedVersion(t *testing.T) {
	var out bytes.Buffer

	writeVersionOutput(&out, "v1.2.3")

	if got := out.String(); got != "v1.2.3\n" {
		t.Fatalf("expected pure version output, got %q", got)
	}
}

func TestResolvedBuildMetadataStripsDirtySuffix(t *testing.T) {
	originalVersion, originalCommit, originalDate := version, commit, date
	originalReadBuildInfo := readBuildInfo
	t.Cleanup(func() {
		version, commit, date = originalVersion, originalCommit, originalDate
		readBuildInfo = originalReadBuildInfo
	})

	version = "dev"
	commit = "none"
	date = "unknown"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v0.1.2+dirty"},
		}, true
	}

	gotVersion, _, _ := resolvedBuildMetadata()
	if gotVersion != "v0.1.2" {
		t.Fatalf("expected dirty suffix to be stripped, got %q", gotVersion)
	}
}
