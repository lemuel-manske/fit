package fit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCreatesInitialConfigurationFile(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "My repository")
	require.NoError(t, err)

	config, err := LoadConfig(dir)
	require.NoError(t, err)

	assert.Equal(t, "My repository", config.RepositoryName)

	assert.True(t, IsUUID(string(config.PeerID)))
	assert.True(t, IsUUID(string(config.RepositoryID)))

	assert.Equal(t, FormatVersion, config.FormatVersion)
}

func TestInitCreatesInitialIndexFile(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "My repository")
	require.NoError(t, err)

	index, err := LoadIndex(dir)
	require.NoError(t, err)

	assert.Equal(t, 0, len(index.Entries))
}
