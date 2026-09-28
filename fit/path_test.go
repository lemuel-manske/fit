package fit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoPathAcceptsNestedRelativePath(t *testing.T) {
	require.NoError(t, ValidateRepoPath("nested/path/to/file.txt"))
}

func TestRepoPathRejectsParentTraversal(t *testing.T) {
	require.Error(t, ValidateRepoPath("../file.txt"))
}

func TestRepoPathRejectsAbsolutePath(t *testing.T) {
	require.Error(t, ValidateRepoPath("/absolute/path/to/file.txt"))
}

func TestRepoPathRejectsDotFit(t *testing.T) {
	require.Error(t, ValidateRepoPath(".."))
}

func TestRepoPathRejectsNonNormalizedPath(t *testing.T) {
	require.Error(t, ValidateRepoPath(".fit/index.json"))
}
