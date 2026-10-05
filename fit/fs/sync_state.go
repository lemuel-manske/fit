package fs

import (
	"errors"
	"os"

	"encoding/json"
)

type SyncHead struct {
	PeerID   PeerID   `json:"peerId"`
	CommitID CommitID `json:"commitId"`
	Complete bool     `json:"complete"`
}

type SyncState struct {
	Heads map[PeerID]SyncHead `json:"heads"`
}

func LoadSyncState(dir string) (*SyncState, error) {
	data, err := os.ReadFile(SyncStatePath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &SyncState{
				Heads: make(map[PeerID]SyncHead),
			}, nil
		}

		return nil, err
	}

	var state SyncState

	if err = json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	if state.Heads == nil {
		state.Heads = make(map[PeerID]SyncHead)
	}

	return &state, nil
}

func WriteSyncState(dir string, state *SyncState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return WriteFileAtomic(
		SyncStatePath(dir),
		data,
		0644,
	)
}

func MarkSyncHead(
	dir string,
	peerID PeerID,
	commitID CommitID,
	complete bool,
) error {
	state, err := LoadSyncState(dir)
	if err != nil {
		return err
	}

	state.Heads[peerID] = SyncHead{
		PeerID:   peerID,
		CommitID: commitID,
		Complete: complete,
	}

	return WriteSyncState(dir, state)
}
