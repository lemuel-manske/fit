package fit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitCreatesInitialConfigurationFile(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "My repository")
	require.NoError(t, err)

	config, err := LoadConfig(dir)
	require.NoError(t, err)

	require.Equal(t, "My repository", config.RepositoryName)

	require.True(t, IsUUID(string(config.PeerID)))
	require.True(t, IsUUID(string(config.RepositoryID)))

	require.Equal(t, FormatVersion, config.FormatVersion)
}

func TestInitCreatesInitialIndexFile(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "My repository")
	require.NoError(t, err)

	index, err := LoadIndex(dir)
	require.NoError(t, err)

	require.Equal(t, 0, len(index.Entries))
}
