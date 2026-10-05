package fs

import (
	"os"
	"testing"

	"path/filepath"

	"github.com/stretchr/testify/require"
)

func TestWriteFileAtomicCreatesFile(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "HEAD")

	require.NoError(t, WriteFileAtomic(path, []byte("commit-id"), 0644))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, []byte("commit-id"), data)

	info, err := os.Stat(path)
	require.NoError(t, err)

	require.Equal(t, os.FileMode(0644), info.Mode().Perm())

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	require.Len(t, entries, 1)
}

func TestWriteFileAtomicReplacesFile(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "index.json")
	require.NoError(t, os.WriteFile(path, []byte("old longer content"), 0644))

	require.NoError(t, WriteFileAtomic(path, []byte("new"), 0600))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, []byte("new"), data)

	info, err := os.Stat(path)
	require.NoError(t, err)

	require.Equal(t, os.FileMode(0600), info.Mode().Perm())

	require.NoError(t, WriteFileAtomic(path, nil, 0600))
	data, err = os.ReadFile(path)
	require.NoError(t, err)

	require.Empty(t, data)
}

func TestWriteFileAtomicRequiresParentDirectory(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "missing", "HEAD")

	err := WriteFileAtomic(path, []byte("commit-id"), 0644)
	require.ErrorIs(t, err, os.ErrNotExist)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	require.Empty(t, entries)
}

func TestWriteFileAtomicCleansUpAfterRenameFailure(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "HEAD")
	child := filepath.Join(path, "existing")

	require.NoError(t, os.Mkdir(path, 0755))

	require.NoError(t, os.WriteFile(child, []byte("original"), 0644))

	require.Error(t, WriteFileAtomic(path, []byte("commit-id"), 0644))

	data, err := os.ReadFile(child)
	require.NoError(t, err)

	require.Equal(t, []byte("original"), data)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	require.Len(t, entries, 1)
	require.Equal(t, "HEAD", entries[0].Name())
}

func TestWriteFileAtomicReplacesSymlink(t *testing.T) {
	dir := t.TempDir()

	target := filepath.Join(dir, "target")
	path := filepath.Join(dir, "HEAD")

	require.NoError(t, os.WriteFile(target, []byte("original"), 0644))
	require.NoError(t, os.Symlink(target, path))

	require.NoError(t, WriteFileAtomic(path, []byte("commit-id"), 0644))

	data, err := os.ReadFile(target)
	require.NoError(t, err)

	require.Equal(t, []byte("original"), data)

	info, err := os.Lstat(path)
	require.NoError(t, err)

	require.True(t, info.Mode().IsRegular())

	data, err = os.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, []byte("commit-id"), data)
}
