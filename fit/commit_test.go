package fit

import (
	"testing"

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

	require.Equal(t, NewCommitID(a), NewCommitID(b))
}

func TestCommitIDIgnoresFileOrder(t *testing.T) {
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

	require.Equal(t, NewCommitID(a), NewCommitID(b))
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
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

	require.Equal(t, NewCommitID(a), NewCommitID(b))
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
}

func TestCommitIDChangesWithContent(t *testing.T) {
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

	require.NotEqual(t, NewCommitID(a), NewCommitID(b))
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

	require.Equal(t, commit.Message, got.Message)
	require.Equal(t, commit.Files, got.Files)
}

func TestCommitStoreRejectsUnsafeManifestPath(t *testing.T) {
	dir := t.TempDir()

	InitCommitStore(dir)

	commit := Commit{
		Files: map[string]Hash{
			"../outside.txt": "abc",
		},
	}

	store := NewCommitStore(dir)

	_, err := store.Put(commit)

	require.Error(t, err)
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

	require.Error(t, err, "Expected error when getting corrupted commit")
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

	require.Equal(t, id1, id2)
}

func TestGetNonExistentCommit(t *testing.T) {
	dir := t.TempDir()

	InitCommitStore(dir)

	store := NewCommitStore(dir)

	_, err := store.Get("nonexistent-id")

	require.Error(t, err, "Expected error when getting non-existent commit")
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

	require.Equal(t, id1, headID)
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

	require.Equal(t, id2, headID)
	require.NotEqual(t, id1, headID)
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

	require.True(t, entry.Delete, "expected a.txt to be staged for deletion")

	_, err = ReadFile(dir, "a.txt")
	require.Error(t, err, "expected a.txt to be removed from working tree")
}

func TestCommitWithMultipleFiles(t *testing.T) {
	dir := t.TempDir()

	Init(dir, "test")

	_, err := WriteFile(dir, "a.txt", "Hello")
	require.NoError(t, err)

	_, err = WriteFile(dir, "b.txt", "World")
	require.NoError(t, err)

	_, err = WriteFile(dir, "a.txt", "Hola")
	require.NoError(t, err)

	err = Add(dir, "a.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "first commit")
	require.NoError(t, err)

	aContent, err := ReadFile(dir, "a.txt")
	require.NoError(t, err)

	bContent, err := ReadFile(dir, "b.txt")
	require.NoError(t, err)

	require.Equal(t, "Hola", string(aContent), "expected a.txt to have latest content in working tree")
	require.Equal(t, "World", string(bContent), "expected b.txt to have latest content in working tree")
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

	require.Equal(t, 1, len(commit2.Parents), "expected second commit to have one parent")
	require.Equal(t, id1, commit2.Parents[0], "expected first commit to be the parent of the second commit")
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

	ok, err := IsAncestor(dir, id3, id1)
	require.NoError(t, err)

	require.True(t, ok, "expected first commit to be an ancestor of the third commit")

	ok, err = IsAncestor(dir, id3, id2)
	require.NoError(t, err)

	require.True(t, ok, "expected second commit to be an ancestor of the third commit")

	ok, err = IsAncestor(dir, id1, id3)
	require.NoError(t, err)

	require.False(t, ok, "expected third commit not to be an ancestor of the first commit")
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

	require.Equal(t, hashA, commit2.Files["a.txt"], "expected second commit to keep file from first commit")
	require.Equal(t, hashB, commit2.Files["b.txt"], "expected second commit to have new file")
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

	require.Empty(t, commit.Parents, "expected first commit to have no parents")
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
	require.False(t, exists, "expected a.txt to be removed from commit after staged deletion")
}
