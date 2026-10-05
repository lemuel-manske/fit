package fs

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

func TestRequireInitializedReturnsErrorWhenConfigMissing(t *testing.T) {
	dir := t.TempDir()

	err := RequireInitialized(dir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "not a FIT repository; run 'fit init' first")
}

func TestRequireInitializedReturnsNoErrorWhenConfigExists(t *testing.T) {
	dir := t.TempDir()

	err := WriteFitDir(dir)
	require.NoError(t, err)

	err = WriteConfig(dir, &Config{
		RepositoryID: "test-repo",
		PeerID:       "test-peer",
	})
	require.NoError(t, err)

	err = RequireInitialized(dir)

	require.NoError(t, err)
}
