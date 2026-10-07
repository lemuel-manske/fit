package repo

import (
	"errors"
	"os"

	"fit/fit/fs"
)

type CommitRelation string

const (
	RelationSame     CommitRelation = "same"
	RelationAhead    CommitRelation = "ahead"
	RelationBehind   CommitRelation = "behind"
	RelationDiverged CommitRelation = "diverged"
	RelationUnknown  CommitRelation = "unknown"
)

func commitRelation(
	dir string,
	local fs.CommitID,
	remote fs.CommitID,
) (CommitRelation, error) {
	if local == remote {
		return RelationSame, nil
	}

	if remote == "" {
		return RelationAhead, nil
	}

	store := fs.NewCommitStore(dir)

	if _, err := store.Get(remote); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RelationUnknown, nil
		}

		return "", err
	}

	if local == "" {
		return RelationBehind, nil
	}

	localIsAncestor, err := fs.IsAncestor(dir, remote, local)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RelationUnknown, nil
		}

		return "", err
	}

	if localIsAncestor {
		return RelationBehind, nil
	}

	remoteIsAncestor, err := fs.IsAncestor(dir, local, remote)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RelationUnknown, nil
		}

		return "", err
	}

	if remoteIsAncestor {
		return RelationAhead, nil
	}

	return RelationDiverged, nil
}
