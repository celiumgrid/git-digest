package git

import "github.com/celiumgrid/git-digest/internal/pathutil"

// NormalizePath expands a leading tilde and returns an absolute path.
func NormalizePath(path string) (string, error) {
	return pathutil.NormalizeUserPath(path)
}
