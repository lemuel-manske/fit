package fit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRmNonExistentFile(t *testing.T) {
	dir := t.TempDir()

	err := Rm(dir, "nonexistent.txt")
	require.Error(t, err)
}

func TestRmEmptyPath(t *testing.T) {
	dir := t.TempDir()

	err := Rm(dir, "")
	require.Error(t, err)
}

func TestRmUntrackedFile(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "untracked.txt", "I am untracked")
	require.NoError(t, err)

	err = Rm(dir, "untracked.txt")
	require.NoError(t, err)

	_, err = ReadFile(dir, "untracked.txt")
	require.Error(t, err)
}

func TestRmAfterAdd(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	content, err := ReadFile(dir, "test.txt")
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", string(content))
}

func TestAddAfterRm(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)
}

func TestRmModifiedFile(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	// modify the file after adding
	_, err = WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	// the file should still exist on disk with modified content
	content, err := ReadFile(dir, "test.txt")
	require.NoError(t, err)

	assert.Equal(t, "Modified content", string(content))
}

func TestRmTwice(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	_, err = ReadFile(dir, "test.txt")
	require.Error(t, err)
}

func TestRmTwiceRemovesFromIndexThenWorkingTree(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteFile(dir, "test.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = WriteFile(dir, "test.txt", "Hola")
	require.NoError(t, err)

	// first rm: remove from index only
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	content, err := ReadFile(dir, "test.txt")
	require.NoError(t, err)
	require.Equal(t, "Hola", string(content))

	// second rm: remove from working tree
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	_, err = ReadFile(dir, "test.txt")
	require.Error(t, err)
}
