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
	fitDir = ".fit"

	configFileName = "config.json"
	indexFileName  = "index.json"
	headFileName   = "HEAD"

	blobDir = "blobs"
	commitsDir = "commits"
)

const (
	ErrCorruptedBlob = "corrupted blob: content does not match its ID"
	ErrCorruptedCommit = "corrupted commit: content does not match its ID"
)

func MakePath(dir string, file string) string {
	return filepath.Join(dir, fitDir, file)
}

func LoadIndex(dir string) (*Index, error) {
	indexPath := MakePath(dir, indexFileName)

	index := &Index{}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, index)
	if err != nil {
		return nil, err
	}

	return index, nil
}

func WriteIndex(dir string, index *Index) error {
	indexPath := MakePath(dir, indexFileName)

	data, err := json.Marshal(index)
	if err != nil {
		return err
	}

	dirPath := filepath.Dir(indexPath)

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(indexPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig(dir string) (*Config, error) {
	configPath := MakePath(dir, configFileName)

	config := &Config{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func WriteConfig(dir string, config *Config) error {
	configPath := MakePath(dir, configFileName)

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	dirPath := filepath.Dir(configPath)

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func WriteHEAD(dir string, commitID CommitID) error {
	path := MakePath(dir, headFileName)

	return os.WriteFile(path, []byte(commitID), 0644)
}

func ReadHEAD(dir string) (CommitID, error) {
	path := MakePath(dir, headFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return CommitID(data), nil
}

func HEADContains(dir string, path string) (bool, error) {
	headID, err := ReadHEAD(dir)
	if err != nil {
		return false, err
	}

	if headID == "" {
		return false, nil
	}

	store := NewCommitStore(dir)

	commit, err := store.Get(headID)
	if err != nil {
		return false, err
	}

	_, exists := commit.Files[path]
	return exists, nil
}

func StagedBlob(dir string, file string) ([]byte, error) {
	index, err := LoadIndex(dir)
	if err != nil {
		return nil, err
	}

	store := NewBlobStore(dir)

	entry, ok := index.Entries[file]
	if !ok {
		err = fmt.Errorf("file %s is not staged", file)

		return nil, err
	}

	content, err := store.Get(Hash(entry.Blob))
	if err != nil {
		return nil, err
	}

	return content, nil
}

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

func InitFsCommitStore(dir string) error {
	commitDirPath := MakePath(dir, commitsDir)

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
	dir := MakePath(s.dir, commitsDir)

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
	if id == "" {
		return Commit{}, fmt.Errorf("commit ID cannot be empty")
	}

	dir := MakePath(s.dir, commitsDir)

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

func NewBlobID(data []byte) Hash {
	sum := sha256.Sum256(data)
	encoded := hex.EncodeToString(sum[:])
	return Hash(encoded)
}

type BlobStore interface {
	Put(data []byte) (Hash, error)
	Get(id Hash) ([]byte, error)
}

type FsBlobStore struct {
	dir string
}

func InitFsBlobStore(dir string) error {
	blobDirPath := MakePath(dir, blobDir)

	// ensure the blobs directory exists
	if err := os.MkdirAll(blobDirPath, 0755); err != nil {
		return err
	}

	return nil
}

func NewBlobStore(dir string) *FsBlobStore {
	return &FsBlobStore{dir: dir}
}

func (s *FsBlobStore) Put(data []byte) (Hash, error) {
	dir := MakePath(s.dir, blobDir)

	id := NewBlobID(data)
	path := filepath.Join(dir, string(id))

	// write the data to the file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return id, nil
}

func (s *FsBlobStore) Get(id Hash) ([]byte, error) {
	dir := MakePath(s.dir, blobDir)

	path := filepath.Join(dir, string(id))

	// read the data from the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// verify that the data matches the expected hash
	if NewBlobID(data) != id {
		err := fmt.Errorf("%s: expected %s, got %s", ErrCorruptedBlob, id, NewBlobID(data))

		return nil, err
	}

	return data, nil
}
