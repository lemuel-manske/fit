package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PeerRef struct {
	PeerID   PeerID   `json:"peerId"`
	CommitID CommitID `json:"commitId"`
}

func WritePeerRef(
	dir string,
	peerID PeerID,
	commitID CommitID,
) error {
	refDir := filepath.Join(dir, ".fit", "refs", "peers")

	if err := os.MkdirAll(refDir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(PeerRef{
		PeerID:   peerID,
		CommitID: commitID,
	})
	if err != nil {
		return err
	}

	return WriteFileAtomic(
		filepath.Join(refDir, string(peerID)+".json"),
		data,
		0644,
	)
}
