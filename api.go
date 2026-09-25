package fit

import (
	"time"

	"encoding/json"
)

const (
	FormatVersion   = 1
	ProtocolVersion = 1
)

type Hash string
type CommitID string
type PeerID string
type RepositoryID string

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
	Name   string `json:"name"`
}

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
	Blob   string `json:"blob,omitempty"`
	Delete bool   `json:"delete,omitempty"`
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
