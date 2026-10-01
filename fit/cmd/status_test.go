package cmd

import (
	"testing"

	utils "fit/fit/test_utils"

	"github.com/stretchr/testify/require"
)

func TestStatusCleanTrackedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.Untracked, "test.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.StagedModified)
}

func TestStatusStagedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.StagedModified, "test.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.Untracked)
}

func TestStatusStagedFileUnmodifiedAfterAdd(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.StagedModified, "test.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.Untracked)
}

func TestStatusStagedFileModifiedAfterAdd(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.Modified, "test.txt")
	require.Contains(t, status.StagedModified, "test.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.Untracked)
}

func TestStatusModifiedTrackedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "test.txt", "Modified content")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.StagedModified, "test.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.Untracked)
}

func TestStatusDeletedTrackedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// unstage
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	// remove
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.Deleted, "test.txt")
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.StagedModified)
	require.Empty(t, status.Untracked)
}

func TestStatusStagedDeletion(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "test.txt", "Hello, World!")
	require.NoError(t, err)

	err = Add(dir, "test.txt")
	require.NoError(t, err)

	_, err = CommitChanges(dir, "Initial commit")
	require.NoError(t, err)

	// remove
	err = Rm(dir, "test.txt")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.StagedDeleted, "test.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedModified)
	require.Empty(t, status.Untracked)
}

func TestStatusUntrackedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "untracked.txt", "I am untracked")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.Untracked, "untracked.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.StagedModified)
}

func TestStatusNestedFile(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, "nested", "file.txt", "Nested file content")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Contains(t, status.Untracked, "nested/file.txt")
	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.StagedModified)
}

func TestStatusIgnoresDotFit(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Init(dir, "test-repo"))

	_, err := utils.WriteFile(dir, ".fit", "ignored.txt", "This should be ignored")
	require.NoError(t, err)

	status, err := GetStatus(dir)
	require.NoError(t, err)

	require.Empty(t, status.Deleted)
	require.Empty(t, status.Modified)
	require.Empty(t, status.StagedDeleted)
	require.Empty(t, status.StagedModified)
	require.Empty(t, status.Untracked)
}
