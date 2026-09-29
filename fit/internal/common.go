package internal

import (
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

func MakePath(dir, file string) string {
	return filepath.Join(dir, fitDir, file)
}

func BlobsPath(dir string) string {
	return MakePath(dir, blobsDir)
}

func BlobPath(dir string, blobID Hash) string {
	return MakePath(dir, filepath.Join(blobsDir, string(blobID)))
}

func CommitsPath(dir string) string {
	return MakePath(dir, commitsDir)
}

func CommitPath(dir string, commitID CommitID) string {
	return MakePath(dir, filepath.Join(commitsDir, string(commitID)))
}

func ConfigPath(dir string) string {
	return MakePath(dir, configFileName)
}

func HEADPath(dir string) string {
	return MakePath(dir, HEADFileName)
}

func IndexPath(dir string) string {
	return MakePath(dir, indexFileName)
}
