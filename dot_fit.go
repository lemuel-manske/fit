package fit

import (
	"fmt"
	"os"

	"encoding/json"
	"path/filepath"
)

const (
	fitDir = ".fit"

	configFileName = "config.json"
	indexFileName  = "index.json"
	headFileName   = "HEAD"
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

func StagedBlob(dir string, file string) ([]byte, error) {
	index, err := LoadIndex(dir)
	if err != nil {
		return nil, err
	}

	store := NewFsBlobStore(dir)

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
