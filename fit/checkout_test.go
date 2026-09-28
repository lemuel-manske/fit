package fit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckout(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// Modify the file after committing
	_, err = WriteFile(dir, "test.txt", "Goodbye, World!")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	content, err := ReadFile(dir, "test.txt")
	require.NoError(t, err)

	require.Equal(t, "Hello, World!", string(content))
}

func TestCheckoutMovesHEAD(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	headID, err := ReadHEAD(dir)
	require.NoError(t, err)

	require.Equal(t, commitID, headID)
}

func TestCheckoutKeepsUntrackedPostCommitFiles(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// Create a new file after committing, and don't commit it
	_, err = WriteFile(dir, "newfile.txt", "This is a new file.")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	// The new file should still exist after checkout
	_, err = ReadFile(dir, "newfile.txt")
	require.NoError(t, err)
}

func TestCheckoutRemovesTrackedPostCommitFiles(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	commitID, err := CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// Create a new file after committing, and commit it
	_, err = WriteFile(dir, "newfile.txt", "This is a new file.")
	require.NoError(t, err)

	err = Add(dir, "newfile.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "Added newfile.txt")
	require.NoError(t, err)

	err = Checkout(dir, commitID)
	require.NoError(t, err)

	// The new file should be removed after checkout
	_, err = ReadFile(dir, "newfile.txt")
	require.Error(t, err)
}
