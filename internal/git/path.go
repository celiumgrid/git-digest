package git

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath expands a leading tilde and returns an absolute path.
func NormalizePath(path string) (string, error) {
	switch {
	case path == "~":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = home
	case strings.HasPrefix(path, "~/"), strings.HasPrefix(path, "~\\"):
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	return filepath.Abs(path)
}
