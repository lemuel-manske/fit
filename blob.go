package fit

import (
	"fmt"
	"os"

	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

const (
	blobDir = "blobs"

	ErrCorruptedBlob = "corrupted blob: content does not match its ID"
)

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

func NewFsBlobStore(dir string) *FsBlobStore {
	return &FsBlobStore{dir: dir}
}

func (s *FsBlobStore) Put(data []byte) (Hash, error) {
	dir := filepath.Join(s.dir, blobDir)

	// ensure the blobs directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	id := NewBlobID(data)
	path := filepath.Join(dir, string(id))

	// write the data to the file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return id, nil
}

func (s *FsBlobStore) Get(id Hash) ([]byte, error) {
	dir := filepath.Join(s.dir, blobDir)

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
