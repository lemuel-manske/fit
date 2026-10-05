package fs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteAndReadMergeHEAD(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	mergeHEAD := CommitID("merge-commit-id")

	require.NoError(t, WriteMergeHEAD(dir, mergeHEAD))

	readMergeHEAD, err := ReadMergeHEAD(dir)
	require.NoError(t, err)

	require.Equal(t, mergeHEAD, readMergeHEAD)
}

func TestWriteAndReadMergeBase(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	mergeBase := CommitID("merge-base-id")

	require.NoError(t, WriteMergeBase(dir, mergeBase))

	readMergeBase, err := ReadMergeBase(dir)
	require.NoError(t, err)

	require.Equal(t, mergeBase, readMergeBase)
}

func TestClearMergeState(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	mergeHEAD := CommitID("merge-commit-id")
	mergeBase := CommitID("merge-base-id")

	require.NoError(t, WriteMergeHEAD(dir, mergeHEAD))
	require.NoError(t, WriteMergeBase(dir, mergeBase))

	require.NoError(t, ClearMergeState(dir))

	_, err := ReadMergeHEAD(dir)
	require.Error(t, err)

	_, err = ReadMergeBase(dir)
	require.Error(t, err)
}

func TestMergeInProgress(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteFitDir(dir))

	mergeHEAD := CommitID("merge-commit-id")
	mergeBase := CommitID("merge-base-id")

	require.NoError(t, WriteMergeHEAD(dir, mergeHEAD))
	require.NoError(t, WriteMergeBase(dir, mergeBase))

	require.True(t, MergeInProgress(dir))

	require.NoError(t, ClearMergeState(dir))

	require.False(t, MergeInProgress(dir))
}
