package fit

import (
	"fmt"
	"os"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
)

const (
	commitsDir = "commits"

	ErrCorruptedCommit = "corrupted commit: content does not match its ID"
)

func NewCommitID(commit Commit) CommitID {
	commit.ID = "" // exclude ID from the hash calculation
	data, _ := json.Marshal(commit)
	sum := sha256.Sum256(data)
	encoded := hex.EncodeToString(sum[:])
	return CommitID(encoded)
}

type CommitStore interface {
	Put(commit Commit) (CommitID, error)
	Get(id CommitID) (Commit, error)
}

type FsCommitStore struct {
	dir string
}

func NewFsCommitStore(dir string) *FsCommitStore {
	return &FsCommitStore{dir: dir}
}

func (s *FsCommitStore) Put(commit Commit) (CommitID, error) {
	dir := filepath.Join(s.dir, commitsDir)

	// ensure the commits directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	id := NewCommitID(commit)
	commit.ID = id // set the ID in the commit struct
	path := filepath.Join(dir, string(id))

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
	dir := filepath.Join(s.dir, commitsDir)
	path := filepath.Join(dir, string(id))

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
		return Commit{}, fmt.Errorf("%s: expected %s, got %s", ErrCorruptedCommit, expectedID, id)
	}

	return commit, nil
}
