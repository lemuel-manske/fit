package internal

import (
	"os"

	"encoding/json"
)

// Index represents the index of files in the repository.
type Index struct {
	Version int                   `json:"version"`
	Entries map[string]IndexEntry `json:"entries"`
}

func (i *Index) UnmarshalJSON(data []byte) error {
	type Alias Index

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if i.Entries == nil {
		i.Entries = make(map[string]IndexEntry)
	}

	return nil
}

type IndexEntry struct {
	Blob     string `json:"blob,omitempty"`
	Conflict bool   `json:"conflict,omitempty"`
	Delete   bool   `json:"delete,omitempty"`
}

func (i *IndexEntry) UnmarshalJSON(data []byte) error {
	type Alias IndexEntry

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return nil
}

func LoadIndex(dir string) (*Index, error) {
	index := &Index{}

	indexPath := IndexPath(dir)

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
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}

	indexPath := IndexPath(dir)

	err = os.WriteFile(indexPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
