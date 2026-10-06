package repo

import "fit/fit/fs"

type CommitRelation string

const (
	RelationSame     CommitRelation = "same"
	RelationAhead    CommitRelation = "ahead"
	RelationBehind   CommitRelation = "behind"
	RelationDiverged CommitRelation = "diverged"
)

func commitRelation(
	dir string,
	local fs.CommitID,
	remote fs.CommitID,
) (CommitRelation, error) {
	if local == remote {
		return RelationSame, nil
	}

	localIsAncestor, err := fs.IsAncestor(dir, remote, local)
	if err != nil {
		return "", err
	}

	if localIsAncestor {
		return RelationBehind, nil
	}

	remoteIsAncestor, err := fs.IsAncestor(dir, local, remote)
	if err != nil {
		return "", err
	}

	if remoteIsAncestor {
		return RelationAhead, nil
	}

	return RelationDiverged, nil
}
