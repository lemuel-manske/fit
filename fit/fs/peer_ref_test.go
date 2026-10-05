package fs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWritePeerRef(t *testing.T) {
	ref := PeerRef{
		PeerID:   PeerID("peer1"),
		CommitID: CommitID("commit1"),
	}

	dir := t.TempDir()

	require.NoError(t, WritePeerRef(dir, ref.PeerID, ref.CommitID))
}
