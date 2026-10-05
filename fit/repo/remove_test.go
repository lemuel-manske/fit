package repo

import (
	"testing"

	"fit/fit/fs"

	"github.com/stretchr/testify/require"
)

func TestRemoveNonExistentFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	err := Remove(dir, "nonexistent.txt")
	require.Error(t, err)
}

func TestRemoveEmptyPath(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	err := Remove(dir, "")
	require.Error(t, err)
}

func TestRemoveUntrackedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "untracked.txt", "I am untracked")
	require.NoError(t, err)

	err = Remove(dir, "untracked.txt")
	require.NoError(t, err)

	_, err = fs.ReadFile(dir, "untracked.txt")
	require.Error(t, err)
}

func TestRemoveAfterAdd(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Remove(dir, "test.txt")
	require.NoError(t, err)

	content, err := fs.ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(content))
}

func TestAddAfterRemove(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Remove(dir, "test.txt")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)
}

func TestRemoveModifiedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	// modify the file after adding
	_, err = fs.WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	err = Remove(dir, "test.txt")
	require.NoError(t, err)

	// the file should still exist on disk with modified content
	content, err := fs.ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Modified content", string(content))
}

func TestRemoveTwiceRemovesFromIndexThenWorkingTree(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "test.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "test.txt", "Hola")
	require.NoError(t, err)

	// first rm: remove from index only
	err = Remove(dir, "test.txt")
	require.NoError(t, err)

	content, err := fs.ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hola", string(content))

	// second rm: remove from working tree
	err = Remove(dir, "test.txt")
	require.NoError(t, err)

	_, err = fs.ReadFile(dir, "test.txt")
	require.Error(t, err)
}

func TestRemoveRejectsPathOutsideRepository(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "test.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	// Attempt to remove a file outside the repository
	err = Remove(dir, "../outside.txt")
	require.Error(t, err)
}
