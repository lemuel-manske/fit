package cmd

import (
	"io"
	"slices"

	"fit/fit/internal"
)

const (
	green = "\033[32m"
	red   = "\033[31m"
	reset = "\033[0m"
)

type Status struct {
	MergeInProgress bool

	StagedModified []string
	StagedDeleted  []string

	Modified  []string
	Deleted   []string
	Untracked []string
}

func (s *Status) IsClean() bool {
	return len(s.StagedModified) == 0 &&
		len(s.StagedDeleted) == 0 &&
		len(s.Modified) == 0 &&
		len(s.Deleted) == 0 &&
		len(s.Untracked) == 0
}

// git-like priting with colors
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
}

func GetStatus(dir string) (*Status, error) {
	status := &Status{
		StagedModified: []string{},
		StagedDeleted:  []string{},
		Modified:       []string{},
		Deleted:        []string{},
		Untracked:      []string{},
	}

	status.MergeInProgress = internal.MergeInProgress(dir)

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return nil, err
	}

	headFiles := map[string]internal.Hash{}

	headID, err := internal.ReadHEAD(dir)
	if err == nil && headID != "" {
		store := internal.NewCommitStore(dir)

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
		stagedHash := internal.Hash(entry.Blob)

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

			workingHash := internal.NewBlobID(content)
			stagedHash := internal.Hash(entry.Blob)

			if workingHash != stagedHash {
				status.Modified = append(status.Modified, path)
			}

			continue
		}

		headHash, tracked := headFiles[path]
		if tracked {
			if internal.NewBlobID(content) != headHash {
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

	slices.Sort(status.StagedModified)
	slices.Sort(status.StagedDeleted)
	slices.Sort(status.Modified)
	slices.Sort(status.Deleted)
	slices.Sort(status.Untracked)

	return status, nil
}
