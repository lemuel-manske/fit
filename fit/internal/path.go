package internal

import (
	"fmt"
	"strings"

	"path/filepath"
)

// ValidateRepoPath checks if the given path is a valid
// relative path within the repository.
func ValidateRepoPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute path")
	}

	clean := filepath.Clean(path)

	if clean == "." || clean == ".." {
		return fmt.Errorf("invalid path")
	}

	if clean != path {
		return fmt.Errorf("path is not normalized")
	}

	if strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes repository")
	}

	if clean == fitDir ||
		strings.HasPrefix(clean, fitDir+string(filepath.Separator)) {

		return fmt.Errorf("path points into internal directory")
	}

	return nil
}
