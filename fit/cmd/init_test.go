package cmd

import (
	"testing"

	"fit/fit/internal"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInitCreatesInitialConfigurationFile(t *testing.T) {
	dir := t.TempDir()

	err := InitNew(dir, "My repository")
	require.NoError(t, err)

	config, err := internal.LoadConfig(dir)
	require.NoError(t, err)

	require.Equal(t, "My repository", config.RepositoryName)

	_, err = uuid.Parse(string(config.PeerID))
	require.NoError(t, err)

	_, err = uuid.Parse(string(config.RepositoryID))
	require.NoError(t, err)

	require.Equal(t, internal.FormatVersion, config.FormatVersion)
}

func TestInitCreatesInitialIndexFile(t *testing.T) {
	dir := t.TempDir()

	err := InitNew(dir, "My repository")
	require.NoError(t, err)

	index, err := internal.LoadIndex(dir)
	require.NoError(t, err)

	require.Equal(t, 0, len(index.Entries))
}
