package fit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddSnapshotsCurrentFileContent(t *testing.T) {
	dir := t.TempDir()

	path := NewFile(dir, "test.txt", "Hello, World!")

	err := Add(dir, "test.txt")
	require.NoError(t, err)

	WriteFile(path, "Goodbye, World!")

	staged, err := StagedBlob(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddTwiceDoesNotDuplicate(t *testing.T) {
	dir := t.TempDir()

	NewFile(dir, "test.txt", "Hello, World!")

	err := Add(dir, "test.txt")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	index := LoadIndex(dir)
	require.Len(t, index.Entries, 1)
}
