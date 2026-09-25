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

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewFsCommitStore(dir)

	id, err := store.Put(commit)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, commit.Message, got.Message)
	assert.Equal(t, commit.Files, got.Files)
}

func TestGetCommitRejectsCorruptedCommit(t *testing.T) {
	dir := t.TempDir()

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewFsCommitStore(dir)

	id, err := store.Put(commit)
	require.NoError(t, err)

	// Corrupt the commit file
	WriteFile(
		dir, "commits", string(id), "{\"message\":\"corrupted commit\"}",
	)

	_, err = store.Get(id)

	assert.Error(t, err, "Expected error when getting corrupted commit")
}

func TestPutCommitTwiceProducesSameID(t *testing.T) {
	dir := t.TempDir()

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewFsCommitStore(dir)

	id1, err := store.Put(commit)
	require.NoError(t, err)

	id2, err := store.Put(commit)
	require.NoError(t, err)

	assert.Equal(t, id1, id2)
}

func TestPutCommitTwiceDoesNotDuplicateIt(t *testing.T) {
	dir := t.TempDir()

	commit := Commit{
		Message: "first commit",
		Files: map[string]Hash{
			"a.txt": "abc123",
		},
	}

	store := NewFsCommitStore(dir)

	id1, err := store.Put(commit)
	require.NoError(t, err)

	id2, err := store.Put(commit)
	require.NoError(t, err)

	assert.Equal(t, id1, id2)

	files, err := ListFiles(dir, "commits")
	require.NoError(t, err)

	assert.Equal(t, 1, len(files), "Expected only one commit file in the store")
}

func TestGetNonExistentCommit(t *testing.T) {
	dir := t.TempDir()

	store := NewFsCommitStore(dir)

	_, err := store.Get("nonexistent-id")

	assert.Error(t, err, "Expected error when getting non-existent commit")
}
