package fs

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

func TestAncestorOfSiblingsIsTheirParent(t *testing.T) {
	dir := t.TempDir()

	a := Commit{
		Message: "first",
	}

	b := Commit{
		Message: "second",
	}

	c := Commit{
		Message: "third",
	}

	store := NewCommitStore(dir)

	aID, err := store.Put(a)
	require.NoError(t, err)

	b.Parents = []CommitID{aID}
	c.Parents = []CommitID{aID}

	bID, err := store.Put(b)
	require.NoError(t, err)

	cID, err := store.Put(c)
	require.NoError(t, err)

	ancestors, err := Ancestor(dir, bID, cID)
	require.NoError(t, err)

	require.Equal(t, aID, ancestors)
}

func TestAncestorWhenOneCommitIsAncestorOfTheOther(t *testing.T) {
	dir := t.TempDir()

	a := Commit{
		Message: "first",
	}

	b := Commit{
		Message: "second",
	}

	store := NewCommitStore(dir)

	aID, err := store.Put(a)
	require.NoError(t, err)

	b.Parents = []CommitID{aID}

	bID, err := store.Put(b)
	require.NoError(t, err)

	ancestors, err := Ancestor(dir, aID, bID)
	require.NoError(t, err)

	require.Equal(t, aID, ancestors)
}

func TestNoCommonAncestor(t *testing.T) {
	dir := t.TempDir()

	a := Commit{
		Message: "first",
	}

	b := Commit{
		Message: "second",
	}

	store := NewCommitStore(dir)

	_, err := store.Put(a)
	require.NoError(t, err)

	_, err = store.Put(b)
	require.NoError(t, err)

	_, err = Ancestor(dir, a.ID, b.ID)
	require.Error(t, err)
}

func TestPutAndGetCommit(t *testing.T) {
	dir := t.TempDir()

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

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewCommitStore(dir)

	id, err := store.Put(commit)
	require.NoError(t, err)

	err = corruptCommit(dir, id, "corrupted commit")
	require.NoError(t, err)

	_, err = store.Get(id)

	require.Error(t, err, "Expected error when getting corrupted commit")
}

func TestPutCommitTwiceProducesSameID(t *testing.T) {
	dir := t.TempDir()

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

	store := NewCommitStore(dir)

	_, err := store.Get("nonexistent-id")

	require.Error(t, err, "Expected error when getting non-existent commit")
}

func corruptCommit(dir string, commitID CommitID, content string) error {
	commitsDir := CommitsPath(dir)

	_, err := WriteFile(commitsDir, string(commitID), content)
	if err != nil {
		return err
	}

	return nil
}
