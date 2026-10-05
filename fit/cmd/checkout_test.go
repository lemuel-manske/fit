package cmd

import (
	"testing"

	"fit/fit/internal"
	utils "fit/fit/test_utils"

	"github.com/stretchr/testify/require"
)

func TestCheckoutMovesHEAD(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	headID, err := internal.ReadHEAD(dir)
	require.NoError(t, err)

	require.Equal(t, commitID, headID)
}

func TestCheckoutKeepsUntrackedPostCommitFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// create a new file after committing, and don't commit it
	_, err = utils.WriteFile(dir, "newfile.txt", "This is a new file.")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	// the new file should still exist after checkout
	_, err = utils.ReadFile(dir, "newfile.txt")
	require.NoError(t, err)
}

func TestCheckoutRemovesTrackedPostCommitFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// create a new file after committing, and commit it
	_, err = utils.WriteFile(dir, "newfile.txt", "This is a new file.")
	require.NoError(t, err)

	err = Add(dir, "newfile.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "Added newfile.txt")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	// the new file should be removed after checkout
	_, err = utils.ReadFile(dir, "newfile.txt")
	require.Error(t, err)
}

func TestCheckoutRejectsModifiedWorkingTree(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// modify the file after committing
	_, err = utils.WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.Error(t, err) // should error due to modified working tree
}

func TestCheckoutRejectsStagedChanges(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// modify the file after committing
	_, err = utils.WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.Error(t, err) // should error due to staged changes
}

func TestCheckoutRejectsUntrackedFileCollision(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// create an untracked file that would collide with a tracked file in the commit
	_, err = utils.WriteFile(dir, "test.txt", "Untracked content")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.Error(t, err) // should error due to untracked file collision
}

func TestCheckoutHandlesFileDirectoryTransition(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, InitNew(dir, "test"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	require.NoError(t, Add(dir, "test.txt"))

	commitA, err := CommitChanges(dir, "file")
	require.NoError(t, err)

	require.NoError(t, Rm(dir, "test.txt"))

	_, err = utils.WriteFile(dir, "test.txt", "nested.txt", "Nested")
	require.NoError(t, err)

	require.NoError(t, Add(dir, "test.txt/nested.txt"))

	commitB, err := CommitChanges(dir, "directory")
	require.NoError(t, err)

	// go back to file state.
	require.NoError(t, Checkout(dir, commitA))

	content, err := utils.ReadFile(dir, "test.txt")
	require.NoError(t, err)
	require.Equal(t, "Hello, World!", string(content))

	// now checkout directory state.
	require.NoError(t, Checkout(dir, commitB))

	content, err = utils.ReadFile(dir, "test.txt/nested.txt")
	require.NoError(t, err)

	require.Equal(t, "Nested", string(content))
}
