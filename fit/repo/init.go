package repo

import (
	"fit/fit/fs"

	"github.com/google/uuid"
)

// Init initializes a new repository in the specified directory with a new configuration.
func Init(repoDir string, repoName string) error {
	config := &fs.Config{
		FormatVersion:  fs.GlobalFormatVersion,
		PeerID:         fs.PeerID(uuid.New().String()),
		RepositoryID:   fs.RepositoryID(uuid.New().String()),
		RepositoryName: repoName,
	}

	return InitWithConfig(repoDir, config)
}

// InitWithConfig initializes an existing repository in the specified directory with the provided configuration.
func InitWithConfig(repoDir string, config *fs.Config) error {
	err := fs.WriteFitDir(repoDir)
	if err != nil {
		return err
	}

	if err = fs.WriteConfig(repoDir, config); err != nil {
		return err
	}

	index := &fs.Index{
		Version: 0,
		Entries: make(map[string]fs.IndexEntry),
	}

	if err = fs.WriteIndex(repoDir, index); err != nil {
		return err
	}

	if err = fs.WriteHEAD(repoDir, ""); err != nil {
		return err
	}

	// create initial .fit/blobs and .fit/commits directories
	_ = fs.NewBlobStore(repoDir)
	_ = fs.NewCommitStore(repoDir)

	return nil
}
