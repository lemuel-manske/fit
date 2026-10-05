package repo

import (
	"testing"

	"fit/fit/fs"

	"github.com/stretchr/testify/require"
)

func TestCommitDoesNotModifyWorkingTree(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "a.txt", "Hola")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	content, err := fs.ReadFile(dir, "a.txt")
	require.NoError(t, err)

	require.Equal(t, "Hola", string(content))
}

func TestFirstCommitMovesHEAD(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	headID, err := fs.ReadHEAD(dir)
	require.NoError(t, err)

	require.Equal(t, id1, headID)
}

func TestSecondCommitMovesHEAD(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	headID, err := fs.ReadHEAD(dir)
	require.NoError(t, err)

	require.Equal(t, id2, headID)
	require.NotEqual(t, id1, headID)
}

func TestFirstCommitContainsStagedFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	store := fs.NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	hash := fs.NewBlobID([]byte("Hello"))

	require.Equal(t, hash, commit.Files["a.txt"])
}

func TestCommitClearsIndex(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	index, err := fs.LoadIndex(dir)
	require.NoError(t, err)

	require.Equal(t, 0, len(index.Entries), "Expected index to be cleared after commit")
}

func TestRmAfterCommit(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	err = Remove(dir, "a.txt")
	require.NoError(t, err)

	index, err := fs.LoadIndex(dir)
	require.NoError(t, err)

	entry, ok := index.Entries["a.txt"]
	require.True(t, ok, "expected a.txt to exist in the index")

	require.True(t, entry.Delete, "expected a.txt to be staged for deletion")

	_, err = fs.ReadFile(dir, "a.txt")
	require.Error(t, err, "expected a.txt to be removed from working tree")
}

func TestCommitWithMultipleFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "a.txt", "Hola")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	aContent, err := fs.ReadFile(dir, "a.txt")
	require.NoError(t, err)

	bContent, err := fs.ReadFile(dir, "b.txt")
	require.NoError(t, err)

	require.Equal(t, "Hola", string(aContent), "expected a.txt to have latest content in working tree")
	require.Equal(t, "World", string(bContent), "expected b.txt to have latest content in working tree")
}

func TestCommitUsesIndexNotWorkingTree(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "a.txt", "Hola")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	store := fs.NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	hash := fs.NewBlobID([]byte("Hello"))

	require.Equal(t, hash, commit.Files["a.txt"])
}

func TestCommitKeepsParentCommits(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	store := fs.NewCommitStore(dir)

	commit2, err := store.Get(id2)
	require.NoError(t, err)

	require.Equal(t, 1, len(commit2.Parents), "expected second commit to have one parent")
	require.Equal(t, id1, commit2.Parents[0], "expected first commit to be the parent of the second commit")
}

func TestCommitWithNAncestors(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "c.txt", "!")
	require.NoError(t, err)

	err = Add(dir, "c.txt")
	require.NoError(t, err)

	id3, err := CommitChanges(dir, "third commit")
	require.NoError(t, err)

	ok, err := fs.IsAncestor(dir, id3, id1)
	require.NoError(t, err)

	require.True(t, ok, "expected first commit to be an ancestor of the third commit")

	ok, err = fs.IsAncestor(dir, id3, id2)
	require.NoError(t, err)

	require.True(t, ok, "expected second commit to be an ancestor of the third commit")

	ok, err = fs.IsAncestor(dir, id1, id3)
	require.NoError(t, err)

	require.False(t, ok, "expected third commit not to be an ancestor of the first commit")
}

func TestSecondCommitKeepsFilesFromParent(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = fs.WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	store := fs.NewCommitStore(dir)

	commit2, err := store.Get(id2)
	require.NoError(t, err)

	hashA := fs.NewBlobID([]byte("Hello"))
	hashB := fs.NewBlobID([]byte("World"))

	require.Equal(t, hashA, commit2.Files["a.txt"], "expected second commit to keep file from first commit")
	require.Equal(t, hashB, commit2.Files["b.txt"], "expected second commit to have new file")
}

func TestFirstCommitHasNoParents(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	store := fs.NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	require.Empty(t, commit.Parents, "expected first commit to have no parents")
}

func TestCommitAppliesStagedDeletion(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test"))

	_, err := fs.WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	// Stage deletion of a.txt
	err = Remove(dir, "a.txt")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "remove a")
	require.NoError(t, err)

	store := fs.NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	_, exists := commit.Files["a.txt"]
	require.False(t, exists, "expected a.txt to be removed from commit after staged deletion")
}
