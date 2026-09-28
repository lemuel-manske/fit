package fit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSameCommitProducesSameID(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.Equal(t, NewCommitID(a), NewCommitID(b))
}

func TestCommitIDDoesNotDependOnFilesMapInsertionOrder(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "first",
		Files: map[string]Hash{
			"b.txt": "bbb",
			"a.txt": "aaa",
		},
	}

	assert.Equal(t, NewCommitID(a), NewCommitID(b))
}

func TestChangingCommitMessageChangesID(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "second",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestChangingCommitFilesChangesID(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "ccc", // changed content
		},
	}

	assert.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestChangingCommitRepositoryIDChangesID(t *testing.T) {
	a := Commit{
		RepositoryID: "repo1",
		Message:      "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		RepositoryID: "repo2", // changed repository ID
		Message:      "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestChangingCommitParentsChangesID(t *testing.T) {
	a := Commit{
		Parents: []CommitID{"parent1"},
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Parents: []CommitID{"parent2"}, // changed parent
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestCommitIDIgnoresExistingID(t *testing.T) {
	a := Commit{
		ID:      "existing-id",
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.Equal(t, NewCommitID(a), NewCommitID(b))
}

func TestCommitIDIsIndependentOfFilesInsertionOrder(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
			"c.txt": "ccc",
		},
	}

	b := Commit{
		Message: "first",
		Files: map[string]Hash{
			"c.txt": "ccc",
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.Equal(t, NewCommitID(a), NewCommitID(b))
}

func TestChangingMessageChangesCommitID(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "second", // changed message
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	assert.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestChangingFileBlobChangesCommitID(t *testing.T) {
	a := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "bbb",
		},
	}

	b := Commit{
		Message: "first",
		Files: map[string]Hash{
			"a.txt": "aaa",
			"b.txt": "ccc", // changed blob
		},
	}

	assert.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestPutAndGetCommit(t *testing.T) {
	dir := t.TempDir()

	InitCommitStore(dir)

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewCommitStore(dir)

	id, err := store.Put(commit)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, commit.Message, got.Message)
	assert.Equal(t, commit.Files, got.Files)
}

func TestGetCommitRejectsCorruptedCommit(t *testing.T) {
	dir := t.TempDir()

	InitCommitStore(dir)

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewCommitStore(dir)

	id, err := store.Put(commit)
	require.NoError(t, err)

	err = CorruptCommit(dir, id, "corrupted commit")
	require.NoError(t, err)

	_, err = store.Get(id)

	assert.Error(t, err, "Expected error when getting corrupted commit")
}

func TestPutCommitTwiceProducesSameID(t *testing.T) {
	dir := t.TempDir()

	InitCommitStore(dir)

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewCommitStore(dir)

	id1, err := store.Put(commit)
	require.NoError(t, err)

	id2, err := store.Put(commit)
	require.NoError(t, err)

	assert.Equal(t, id1, id2)
}

func TestGetNonExistentCommit(t *testing.T) {
	dir := t.TempDir()

	InitCommitStore(dir)

	store := NewCommitStore(dir)

	_, err := store.Get("nonexistent-id")

	assert.Error(t, err, "Expected error when getting non-existent commit")
}

func TestCommitDoesNotModifyWorkingTree(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = WriteFile(dir, "a.txt", "Hola")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	content, err := ReadFile(dir, "a.txt")
	require.NoError(t, err)

	require.Equal(t, "Hola", string(content))
}

func TestFirstCommitMovesHEAD(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	headID, err := ReadHEAD(dir)
	require.NoError(t, err)

	assert.Equal(t, id1, headID)
}

func TestSecondCommitMovesHEAD(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	headID, err := ReadHEAD(dir)
	require.NoError(t, err)

	assert.Equal(t, id2, headID)
	assert.NotEqual(t, id1, headID)
}

func TestFirstCommitContainsStagedFiles(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	store := NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	hash := NewBlobID([]byte("Hello"))

	require.Equal(t, hash, commit.Files["a.txt"])
}

func TestCommitClearsIndex(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	index, err := LoadIndex(dir)
	require.NoError(t, err)

	require.Equal(t, 0, len(index.Entries), "Expected index to be cleared after commit")
}

func TestRmAfterCommit(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "test")
	require.NoError(t, err)

	_, err = WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	err = Rm(dir, "a.txt")
	require.NoError(t, err)

	index, err := LoadIndex(dir)
	require.NoError(t, err)

	entry, ok := index.Entries["a.txt"]
	require.True(t, ok, "expected a.txt to exist in the index")

	assert.True(t, entry.Delete, "expected a.txt to be staged for deletion")

	_, err = ReadFile(dir, "a.txt")
	require.Error(t, err, "expected a.txt to be removed from working tree")
}

func TestCommitUsesIndexNotWorkingTree(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = WriteFile(dir, "a.txt", "Hola")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	store := NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	hash := NewBlobID([]byte("Hello"))

	require.Equal(t, hash, commit.Files["a.txt"])
}

func TestCommitKeepsParentCommits(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	store := NewCommitStore(dir)

	commit2, err := store.Get(id2)
	require.NoError(t, err)

	assert.Equal(t, 1, len(commit2.Parents), "expected second commit to have one parent")
	assert.Equal(t, id1, commit2.Parents[0], "expected first commit to be the parent of the second commit")
}

func TestCommitWithNAncestors(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id1, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	_, err = WriteFile(dir, "c.txt", "!")
	require.NoError(t, err)

	err = Add(dir, "c.txt")
	require.NoError(t, err)

	id3, err := CommitChanges(dir, "third commit")
	require.NoError(t, err)

	id1AncestorID3, err := Ancestor(dir, id3, id1)
	require.NoError(t, err)

	assert.True(t, id1AncestorID3, "expected first commit to be an ancestor of the third commit")

	id2AncestorID3, err := Ancestor(dir, id3, id2)
	require.NoError(t, err)

	assert.True(t, id2AncestorID3, "expected second commit to be an ancestor of the third commit")

	id3AncestorID1, err := Ancestor(dir, id1, id3)
	require.NoError(t, err)

	assert.False(t, id3AncestorID1, "expected third commit not to be an ancestor of the first commit")
}

func TestSecondCommitKeepsFilesFromParent(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	_, err = WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	err = Add(dir, "b.txt")
	require.NoError(t, err)

	id2, err := CommitChanges(dir, "second commit")
	require.NoError(t, err)

	store := NewCommitStore(dir)

	commit2, err := store.Get(id2)
	require.NoError(t, err)

	hashA := NewBlobID([]byte("Hello"))
	hashB := NewBlobID([]byte("World"))

	assert.Equal(t, hashA, commit2.Files["a.txt"], "expected second commit to keep file from first commit")
	assert.Equal(t, hashB, commit2.Files["b.txt"], "expected second commit to have new file")
}

func TestFirstCommitHasNoParents(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "first commit")
	require.NoError(t, err)

	store := NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	assert.Empty(t, commit.Parents, "expected first commit to have no parents")
}

func TestCommitAppliesStagedDeletion(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	// Stage deletion of a.txt
	err = Rm(dir, "a.txt")
	require.NoError(t, err)

	id, err := CommitChanges(dir, "remove a")
	require.NoError(t, err)

	store := NewCommitStore(dir)

	commit, err := store.Get(id)
	require.NoError(t, err)

	_, exists := commit.Files["a.txt"]
	assert.False(t, exists, "expected a.txt to be removed from commit after staged deletion")
}
