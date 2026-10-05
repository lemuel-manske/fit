package cmd

import (
	"os"

	"path/filepath"
)

const (
	RemoteURL = "amqp://guest:guest@localhost:5672"
)

func workingTreeFiles(root string) (map[string][]byte, error) {
	files := make(map[string][]byte)

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if rel == ".fit" {
				return filepath.SkipDir
			}

			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files[filepath.ToSlash(rel)] = content

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
