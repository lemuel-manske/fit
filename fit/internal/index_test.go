package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteAndReadIndex(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	index := Index{
		Version: 1,
		Entries: map[string]IndexEntry{
			"file1.txt": {
				Blob: "hash1",
			},
			"file2.txt": {
				Blob: "hash2",
			},
		},
	}

	require.NoError(t, WriteIndex(dir, &index))

	readIndex, err := LoadIndex(dir)
	require.NoError(t, err)

	require.Equal(t, &index, readIndex)
}
