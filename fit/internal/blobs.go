package internal

import (
	"fmt"
	"os"

	"crypto/sha256"
	"encoding/hex"
)

const (
	errCorruptedBlob = "corrupted blob"
)

type BlobStore interface {
	Put(data []byte) (Hash, error)
	Get(id Hash) ([]byte, error)
}

type FsBlobStore struct {
	dir string
}

func NewBlobID(data []byte) Hash {
	sum := sha256.Sum256(data)
	encoded := hex.EncodeToString(sum[:])
	return Hash(encoded)
}

func NewBlobStore(dir string) *FsBlobStore {
	blobDirPath := BlobsPath(dir)

	// ensure the blobs directory exists
	if err := os.MkdirAll(blobDirPath, 0755); err != nil {
		panic(err)
	}

	return &FsBlobStore{dir: dir}
}

func (s *FsBlobStore) Put(data []byte) (Hash, error) {
	id := NewBlobID(data)

	path := BlobPath(s.dir, id)

	// write the data to the file
	err := WriteFileAtomic(path, data, 0644)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *FsBlobStore) Get(id Hash) ([]byte, error) {
	path := BlobPath(s.dir, id)

	// read the data from the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// verify that the data matches the expected hash
	if NewBlobID(data) != id {
		err := fmt.Errorf("%s: expected %s, got %s", errCorruptedBlob, id, NewBlobID(data))

		return nil, err
	}

	return data, nil
}

func StagedBlob(dir string, file string) ([]byte, error) {
	index, err := LoadIndex(dir)
	if err != nil {
		return nil, err
	}

	entry, ok := index.Entries[file]
	if !ok {
		return nil, fmt.Errorf("file %s is not staged", file)
	}

	if entry.Delete {
		return nil, fmt.Errorf("file %s is staged for deletion", file)
	}

	store := NewBlobStore(dir)

	content, err := store.Get(Hash(entry.Blob))
	if err != nil {
		return nil, err
	}

	return content, nil
}
