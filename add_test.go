package fit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddNonExistentFile(t *testing.T) {
	dir := t.TempDir()

	err := Add(dir, "nonexistent.txt")
	require.Error(t, err)
}

func TestAddEmptyPath(t *testing.T) {
	dir := t.TempDir()

	err := Add(dir, "")
	require.Error(t, err)
}

func TestUpdateFileBlobStaysSame(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	// update the file content
	WriteFile(dir, "test.txt", "Goodbye, World!")

	staged, err := StagedBlob(dir, "test.txt")
	require.NoError(t, err)

	// the staged blob should still be the original content
	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddSameFileTwice(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	staged, err := StagedBlob(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddFile(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = WriteFile(dir, "test.txt", "Goodbye, World!")
	require.NoError(t, err)

	staged, err := StagedBlob(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddDeepFile(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "subdir", "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "subdir/test.txt")
	require.NoError(t, err)

	staged, err := StagedBlob(dir, "subdir/test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddVeryDeepFile(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "subdir", "subdir2", "subdir3", "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "subdir/subdir2/subdir3/test.txt")
	require.NoError(t, err)

	staged, err := StagedBlob(dir, "subdir/subdir2/subdir3/test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddFileRelativeToRepoRoot(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	subdir := filepath.Join(dir, "subdir")

	err = Add(subdir, "../test.txt")
	require.NoError(t, err)

	staged, err := StagedBlob(subdir, "../test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(staged))
}

func TestAddDir(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "subdir", "test1.txt", "Hello, World!")
	require.NoError(t, err)

	_, err = WriteFile(dir, "subdir", "test2.txt", "Goodbye, World!")
	require.NoError(t, err)

	err = Add(dir, "subdir")
	require.NoError(t, err)

	staged1, err := StagedBlob(dir, "subdir/test1.txt")
	require.NoError(t, err)
	require.Equal(t, "Hello, World!", string(staged1))

	staged2, err := StagedBlob(dir, "subdir/test2.txt")
	require.NoError(t, err)
	require.Equal(t, "Goodbye, World!", string(staged2))
}
