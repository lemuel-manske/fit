package repo

import (
	"fmt"
	"io"
	"slices"

	"fit/fit/fs"
)

const (
	green = "\033[32m"
	red   = "\033[31m"
	reset = "\033[0m"
)

type RemoteStatus struct {
	PeerID   fs.PeerID
	Head     fs.CommitID
	Relation CommitRelation
}

type Status struct {
	MergeInProgress bool

	StagedModified []string
	StagedDeleted  []string

	Modified  []string
	Deleted   []string
	Untracked []string

	Remotes []RemoteStatus
}

func (s *Status) IsClean() bool {
	return !s.MergeInProgress &&
		len(s.StagedModified) == 0 &&
		len(s.StagedDeleted) == 0 &&
		len(s.Modified) == 0 &&
		len(s.Deleted) == 0 &&
		len(s.Untracked) == 0
}

// Print works as a git-like status output
func (s *Status) Print(writer io.Writer) {
	if s.MergeInProgress {
		_, _ = writer.Write([]byte("Merge in progress...\n"))
	}

	if len(s.StagedModified) > 0 {
		_, _ = writer.Write([]byte(green + "Staged modified files:\n" + reset))
		for _, file := range s.StagedModified {
			_, _ = writer.Write([]byte("  " + file + "\n"))
		}
	}

	if len(s.StagedDeleted) > 0 {
		_, _ = writer.Write([]byte(red + "Staged deleted files:\n" + reset))
		for _, file := range s.StagedDeleted {
			_, _ = writer.Write([]byte("  " + file + "\n"))
		}
	}

	if len(s.Modified) > 0 {
		_, _ = writer.Write([]byte(red + "Modified files:\n" + reset))
		for _, file := range s.Modified {
			_, _ = writer.Write([]byte("  " + file + "\n"))
		}
	}

	if len(s.Deleted) > 0 {
		_, _ = writer.Write([]byte(red + "Deleted files:\n" + reset))
		for _, file := range s.Deleted {
			_, _ = writer.Write([]byte("  " + file + "\n"))
		}
	}

	if len(s.Untracked) > 0 {
		_, _ = writer.Write([]byte(red + "Untracked files:\n" + reset))
		for _, file := range s.Untracked {
			_, _ = writer.Write([]byte("  " + file + "\n"))
		}
	}

	if s.IsClean() {
		_, _ = writer.Write([]byte(green + "Working tree clean\n" + reset))
	}

	if len(s.Remotes) == 0 {
		return
	}

	_, _ = writer.Write([]byte("\nRemote heads:\n"))

	for _, remote := range s.Remotes {
		_, _ = fmt.Fprintf(
			writer,
			"  %s  %s  %s\n",
			remote.PeerID,
			remote.Relation,
			remote.Head,
		)

		switch remote.Relation {
		case RelationBehind, RelationDiverged:
			_, _ = fmt.Fprintf(writer, "    run: fit merge %s\n", remote.Head)
		case RelationUnknown:
			_, _ = writer.Write([]byte("    run: fit sync\n"))
		}
	}
}

func GetStatus(dir string) (*Status, error) {
	err := fs.RequireInitialized(dir)
	if err != nil {
		return nil, err
	}

	status := &Status{
		StagedModified: []string{},
		StagedDeleted:  []string{},
		Modified:       []string{},
		Deleted:        []string{},
		Untracked:      []string{},
		Remotes:        []RemoteStatus{},
	}

	status.MergeInProgress = fs.MergeInProgress(dir)

	index, err := fs.LoadIndex(dir)
	if err != nil {
		return nil, err
	}

	headFiles := map[string]fs.Hash{}

	headID, err := fs.ReadHEAD(dir)
	if err != nil {
		return nil, err
	}

	if headID != "" {
		store := fs.NewCommitStore(dir)

		head, err := store.Get(headID)
		if err != nil {
			return nil, err
		}

		headFiles = head.Files
	}

	workingFiles, err := workingTreeFiles(dir)
	if err != nil {
		return nil, err
	}

	for path, entry := range index.Entries {
		if entry.Delete {
			status.StagedDeleted = append(status.StagedDeleted, path)
			continue
		}

		headHash, tracked := headFiles[path]
		stagedHash := fs.Hash(entry.Blob)

		if !tracked || headHash != stagedHash {
			status.StagedModified = append(status.StagedModified, path)
		}
	}

	for path, content := range workingFiles {
		entry, staged := index.Entries[path]

		if staged {
			if entry.Delete {
				status.Modified = append(status.Modified, path)
				continue
			}

			workingHash := fs.NewBlobID(content)
			stagedHash := fs.Hash(entry.Blob)

			if workingHash != stagedHash {
				status.Modified = append(status.Modified, path)
			}

			continue
		}

		headHash, tracked := headFiles[path]
		if tracked {
			if fs.NewBlobID(content) != headHash {
				status.Modified = append(status.Modified, path)
			}

			continue
		}

		status.Untracked = append(status.Untracked, path)
	}

	for path, entry := range index.Entries {
		if entry.Delete {
			continue
		}

		if _, exists := workingFiles[path]; !exists {
			status.Deleted = append(status.Deleted, path)
		}
	}

	for path := range headFiles {
		if _, staged := index.Entries[path]; staged {
			continue
		}

		if _, exists := workingFiles[path]; !exists {
			status.Deleted = append(status.Deleted, path)
		}
	}

	refs, err := fs.ReadPeerRefs(dir)
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		relation, err := commitRelation(dir, headID, ref.CommitID)
		if err != nil {
			return nil, err
		}

		status.Remotes = append(status.Remotes, RemoteStatus{
			PeerID:   ref.PeerID,
			Head:     ref.CommitID,
			Relation: relation,
		})
	}

	slices.Sort(status.StagedModified)
	slices.Sort(status.StagedDeleted)
	slices.Sort(status.Modified)
	slices.Sort(status.Deleted)
	slices.Sort(status.Untracked)

	slices.SortFunc(status.Remotes, func(a, b RemoteStatus) int {
		if a.PeerID < b.PeerID {
			return -1
		}

		if a.PeerID > b.PeerID {
			return 1
		}

		return 0
	})

	return status, nil
}
