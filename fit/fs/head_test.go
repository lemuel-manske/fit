package fs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteAndReadHEAD(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	require.NoError(t, WriteHEAD(dir, "commit-id"))

	headID, err := ReadHEAD(dir)
	require.NoError(t, err)

	require.Equal(t, CommitID("commit-id"), headID)
}

func TestHEADContains(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	commit := Commit{
		Message: "Initial commit",
		Files: map[string]Hash{
			"file.txt": "hash",
		},
	}

	store := NewCommitStore(dir)

	commitID, err := store.Put(commit)
	require.NoError(t, err)

	require.NoError(t, WriteHEAD(dir, commitID))

	exists, err := HEADContains(dir, "file.txt")
	require.NoError(t, err)

	require.True(t, exists)

	exists, err = HEADContains(dir, "nonexistent.txt")
	require.NoError(t, err)

	require.False(t, exists)
}
