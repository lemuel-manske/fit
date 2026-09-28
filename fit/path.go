package fit

import (
	"fmt"
	"strings"

	"path/filepath"
)

// ValidateRepoPath checks if the given path is a valid relative path within the repository.
func ValidateRepoPath(path string) error {
	if IsEmpty(path) {
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

	if clean == FirDir ||
		strings.HasPrefix(clean, FirDir+string(filepath.Separator)) {
		return fmt.Errorf("path points into .fit")
	}

	return nil
}

// ResolveRepoPath resolves a relative path within the repository to an absolute path on the filesystem.
func ResolveRepoPath(root, path string) (string, error) {
	if err := ValidateRepoPath(path); err != nil {
		return "", err
	}

	return filepath.Join(root, path), nil
}
