package internal

import (
	"os"

	"encoding/json"
)

// Config keeps track of peer metadata and repository information.
type Config struct {
	FormatVersion  int          `json:"formatVersion"`
	RepositoryID   RepositoryID `json:"repositoryId"`
	PeerID         PeerID       `json:"peerId"`
	RepositoryName string       `json:"repositoryName"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return nil
}

func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

func LoadConfig(dir string) (*Config, error) {
	configPath := ConfigPath(dir)

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
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configPath := ConfigPath(dir)

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
