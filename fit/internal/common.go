package internal

import (
	"errors"
	"fmt"
	"os"

	"path/filepath"
)

const (
	FormatVersion = 1

	fitDir = ".fit"

	HEADFileName      = "HEAD"
	blobsDir          = "blobs"
	commitsDir        = "commits"
	configFileName    = "config.json"
	indexFileName     = "index.json"
	mergeHEADFileName = "MERGE_HEAD"
	mergeBaseFileName = "MERGE_BASE"
)

type Hash string
type CommitID string
type PeerID string
type RepositoryID string

func fitPath(dir, file string) string {
	return filepath.Join(dir, fitDir, file)
}

func WriteFitDir(dir string) error {
	return os.MkdirAll(filepath.Join(dir, fitDir), 0755)
}

func RequireInitialized(dir string) error {
	configPath := filepath.Join(dir, ".fit", "config.json")

	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not a FIT repository; run 'fit init' first")
		}

		return err
	}

	return nil
}

func BlobsPath(dir string) string {
	return fitPath(dir, blobsDir)
}

func BlobPath(dir string, blobID Hash) string {
	return fitPath(dir, filepath.Join(blobsDir, string(blobID)))
}

func CommitsPath(dir string) string {
	return fitPath(dir, commitsDir)
}

func CommitPath(dir string, commitID CommitID) string {
	return fitPath(dir, filepath.Join(commitsDir, string(commitID)))
}

func ConfigPath(dir string) string {
	return fitPath(dir, configFileName)
}

func HEADPath(dir string) string {
	return fitPath(dir, HEADFileName)
}

func MergeHEADPath(dir string) string {
	return fitPath(dir, mergeHEADFileName)
}

func MergeBasePath(dir string) string {
	return fitPath(dir, mergeBaseFileName)
}

func IndexPath(dir string) string {
	return fitPath(dir, indexFileName)
}
