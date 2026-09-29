package cmd

import (
	"testing"

	utils "fit/fit/test_utils"

	"github.com/stretchr/testify/require"
)

func TestRmNonExistentFile(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test-repo")

	err := Rm(dir, "nonexistent.txt")
	require.Error(t, err)
}

func TestRmEmptyPath(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test-repo")

	err := Rm(dir, "")
	require.Error(t, err)
}

func TestRmUntrackedFile(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test-repo")

	_, err := utils.WriteFile(dir, "untracked.txt", "I am untracked")
	require.NoError(t, err)

	err = Rm(dir, "untracked.txt")
	require.NoError(t, err)

	_, err = utils.ReadFile(dir, "untracked.txt")
	require.Error(t, err)
}

func TestRmAfterAdd(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test-repo")

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	content, err := utils.ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(content))
}

func TestAddAfterRm(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test-repo")

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
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

	Init(dir, "test-repo")

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	// modify the file after adding
	_, err = utils.WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	// the file should still exist on disk with modified content
	content, err := utils.ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Modified content", string(content))
}

func TestRmTwiceRemovesFromIndexThenWorkingTree(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test-repo")

	_, err := utils.WriteFile(dir, "test.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "test.txt", "Hola")
	require.NoError(t, err)

	// first rm: remove from index only
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	content, err := utils.ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hola", string(content))

	// second rm: remove from working tree
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	_, err = utils.ReadFile(dir, "test.txt")
	require.Error(t, err)
}
