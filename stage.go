package fit

import (
	"fmt"
)

func StagedBlob(dir string, file string) ([]byte, error) {
	index := LoadIndex(dir)

	store := NewFsBlobStore(dir)

	entry, ok := index.Entries[file]
	if !ok {
		err := fmt.Errorf("file %s is not staged", file)

		return nil, err
	}

	content, err := store.Get(Hash(entry.Blob))
	if err != nil {
		return nil, err
	}

	return content, nil
}
