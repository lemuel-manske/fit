package fit

import (
	"os"

	"encoding/json"
	"path/filepath"
)

func LoadIndex(dir string) *Index {
	indexPath := filepath.Join(dir, "index.json")
	index := &Index{}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		// if the file doesn't exist, return an empty index
		if os.IsNotExist(err) {
			index.Entries = make(map[string]IndexEntry)
			return index
		}

		panic(err)
	}

	err = json.Unmarshal(data, index)
	if err != nil {
		panic(err)
	}

	return index
}

func WriteIndex(dir string, index *Index) {
	indexPath := filepath.Join(dir, "index.json")

	data, err := json.Marshal(index)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(indexPath, data, 0644)
	if err != nil {
		panic(err)
	}
}

func ReadFile(dir, file string) ([]byte, error) {
	path := filepath.Join(dir, file)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}
