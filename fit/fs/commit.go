package fs

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

func NewCommitStore(dir string) *FsCommitStore {
	commitDirPath := CommitsPath(dir)

	// ensure the commits directory exists
	if err := os.MkdirAll(commitDirPath, 0755); err != nil {
		panic(err)
	}

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

	err = WriteFileAtomic(path, data, 0644)
	if err != nil {
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

func Ancestor(dir string, commits ...CommitID) (CommitID, error) {
	if len(commits) < 2 {
		return "", fmt.Errorf("at least two commits are required")
	}

	store := NewCommitStore(dir)

	distances := make([]map[CommitID]int, len(commits))

	for i, id := range commits {
		d, err := ancestorDistances(store, id)
		if err != nil {
			return "", err
		}

		distances[i] = d
	}

	var best CommitID
	bestScore := int(^uint(0) >> 1)
	found := false

	for candidate, d0 := range distances[0] {
		score := d0
		common := true

		for i := 1; i < len(distances); i++ {
			d, ok := distances[i][candidate]
			if !ok {
				common = false
				break
			}

			score += d
		}

		if !common {
			continue
		}

		if !found || score < bestScore {
			best = candidate
			bestScore = score
			found = true
		} else if score == bestScore && candidate != best {
			return "", fmt.Errorf("multiple closest common ancestors")
		}
	}

	if !found {
		return "", fmt.Errorf("no common ancestor found")
	}

	return best, nil
}

func ancestorDistances(
	store CommitStore,
	start CommitID,
) (map[CommitID]int, error) {
	dist := map[CommitID]int{start: 0}
	queue := []CommitID{start}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		commit, err := store.Get(current)
		if err != nil {
			return nil, err
		}

		for _, parent := range commit.Parents {
			nextDistance := dist[current] + 1

			old, seen := dist[parent]
			if seen && old <= nextDistance {
				continue
			}

			dist[parent] = nextDistance
			queue = append(queue, parent)
		}
	}

	return dist, nil
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
