package repo

import (
	"testing"

	"fit/fit/fs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInitCreatesInitialConfigurationFile(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "My repository")
	require.NoError(t, err)

	config, err := fs.LoadConfig(dir)
	require.NoError(t, err)

	require.Equal(t, "My repository", config.RepositoryName)

	_, err = uuid.Parse(string(config.PeerID))
	require.NoError(t, err)

	_, err = uuid.Parse(string(config.RepositoryID))
	require.NoError(t, err)

	require.Equal(t, fs.GlobalFormatVersion, config.FormatVersion)
}

func TestInitCreatesInitialIndexFile(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "My repository")
	require.NoError(t, err)

	index, err := fs.LoadIndex(dir)
	require.NoError(t, err)

	require.Equal(t, 0, len(index.Entries))
}
