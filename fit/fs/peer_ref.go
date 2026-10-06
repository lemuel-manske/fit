package fs

import (
	"errors"
	"os"

	"encoding/json"
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
	refDir := PeerRefsPath(dir)

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

func ReadPeerRefs(dir string) ([]PeerRef, error) {
	refDir := PeerRefsPath(dir)

	entries, err := os.ReadDir(refDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []PeerRef{}, nil
		}

		return nil, err
	}

	refs := make([]PeerRef, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(refDir, entry.Name()))
		if err != nil {
			return nil, err
		}

		var ref PeerRef
		if err = json.Unmarshal(data, &ref); err != nil {
			return nil, err
		}

		refs = append(refs, ref)
	}

	return refs, nil
}
