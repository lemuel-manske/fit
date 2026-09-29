package internal

import (
	"fmt"
	"os"
	"time"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	errCorruptedCommit = "corrupted commit"
)

type Commit struct {
	ID           CommitID        `json:"id"`
	RepositoryID RepositoryID    `json:"repositoryId"`
	Parents      []CommitID      `json:"parents"`
	Author       Author          `json:"author"`
	Timestamp    time.Time       `json:"timestamp"`
	Message      string          `json:"message"`
	Files        map[string]Hash `json:"files"`
}

type Author struct {
	PeerID PeerID `json:"peerId"`
	Name   string `json:"name"`
}

type CommitStore interface {
	Put(commit Commit) (CommitID, error)
	Get(id CommitID) (Commit, error)
}

type FsCommitStore struct {
	dir string
}

func NewCommitID(commit Commit) CommitID {
	commit.ID = "" // exclude ID from the hash calculation
	data, _ := json.Marshal(commit)
	sum := sha256.Sum256(data)
	encoded := hex.EncodeToString(sum[:])
	return CommitID(encoded)
}

func InitCommitStore(dir string) error {
	commitDirPath := CommitsPath(dir)

	// ensure the commits directory exists
	if err := os.MkdirAll(commitDirPath, 0755); err != nil {
		return err
	}

	return nil
}

func NewCommitStore(dir string) *FsCommitStore {
	return &FsCommitStore{dir: dir}
}

func (s *FsCommitStore) Put(commit Commit) (CommitID, error) {
	for path := range commit.Files {
		err := ValidateRepoPath(path)

		if err != nil {
			return "", fmt.Errorf(
				"%s: invalid path %q: %w",
				errCorruptedCommit,
				path,
				err,
			)
		}
	}

	id := NewCommitID(commit)
	commit.ID = id // set the ID in the commit struct

	path := CommitPath(s.dir, id)

	// write the commit to the file
	data, err := json.Marshal(commit)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return id, nil
}

func (s *FsCommitStore) Get(id CommitID) (Commit, error) {
	if id == "" {
		return Commit{}, fmt.Errorf("commit ID cannot be empty")
	}

	path := CommitPath(s.dir, id)

	// read the commit from the file
	data, err := os.ReadFile(path)
	if err != nil {
		return Commit{}, err
	}

	var commit Commit
	if err := json.Unmarshal(data, &commit); err != nil {
		return Commit{}, err
	}

	expectedID := NewCommitID(commit)
	if expectedID != id {
		return Commit{}, fmt.Errorf("%s: expected %s, got %s", errCorruptedCommit, expectedID, id)
	}

	return commit, nil
}

func IsAncestor(dir string, descendant, ancestor CommitID) (bool, error) {
	if ancestor == descendant {
		return true, nil
	}

	store := NewCommitStore(dir)

	visited := map[CommitID]bool{}
	stack := []CommitID{descendant}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}
		visited[current] = true

		commit, err := store.Get(current)
		if err != nil {
			return false, err
		}

		for _, parent := range commit.Parents {
			if parent == ancestor {
				return true, nil
			}

			stack = append(stack, parent)
		}
	}

	return false, nil
}

func CurrentCommit(dir string) (Commit, error) {
	headID, err := ReadHEAD(dir)
	if err != nil {
		return Commit{}, err
	}

	if headID == "" {
		return Commit{}, nil
	}

	store := NewCommitStore(dir)

	commit, err := store.Get(headID)
	if err != nil {
		return Commit{}, err
	}

	return commit, nil
}
