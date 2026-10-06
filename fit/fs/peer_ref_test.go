package fs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAndWritePeerRef(t *testing.T) {
	ref := PeerRef{
		PeerID:   PeerID("peer1"),
		CommitID: CommitID("commit1"),
	}

	dir := t.TempDir()

	require.NoError(t, WritePeerRef(dir, ref.PeerID, ref.CommitID))

	refs, err := ReadPeerRefs(dir)
	require.NoError(t, err)

	require.Len(t, refs, 1)
	require.Equal(t, ref, refs[0])
}

func TestReadPeerRefsEmpty(t *testing.T) {
	dir := t.TempDir()

	refs, err := ReadPeerRefs(dir)
	require.NoError(t, err)

	require.Len(t, refs, 0)
}
