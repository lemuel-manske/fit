package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteAndLoadConfig(t *testing.T) {
	config := &Config{
		RepositoryName: "My repository",
		PeerID:         PeerID("peer-id"),
		RepositoryID:   RepositoryID("repo-id"),
		FormatVersion:  FormatVersion,
	}

	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	require.NoError(t, WriteConfig(dir, config))

	config, err := LoadConfig(dir)
	require.NoError(t, err)

	require.Equal(t, config, config)
}

func TestLoadConfigNonExistent(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadConfig(dir)
	require.Error(t, err)
}

func TestLoadConfigInvalidVersion(t *testing.T) {
	dir := t.TempDir()

	config := &Config{
		RepositoryName: "My repository",
		PeerID:         PeerID("peer-id"),
		RepositoryID:   RepositoryID("repo-id"),
		FormatVersion:  2,
	}

	require.NoError(t, WriteFitDir(dir))
	require.NoError(t, WriteConfig(dir, config))

	_, err := LoadConfig(dir)
	require.Error(t, err)
}
