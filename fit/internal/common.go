package internal

import (
	"os"

	"path/filepath"
)

const (
	FormatVersion = 1

	fitDir = ".fit"

	blobsDir       = "blobs"
	commitsDir     = "commits"
	configFileName = "config.json"
	HEADFileName   = "HEAD"
	indexFileName  = "index.json"
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

func IndexPath(dir string) string {
	return fitPath(dir, indexFileName)
}
