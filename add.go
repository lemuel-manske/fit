package fit

func Add(dir string, file string) error {
	index := LoadIndex(dir)

	store := NewFsBlobStore(dir)

	fileContent, err := ReadFile(dir, file)
	if err != nil {
		return err
	}

	blobID, err := store.Put(fileContent)
	if err != nil {
		return err
	}

	index.Entries[file] = IndexEntry{
		Blob: string(blobID),
	}

	WriteIndex(dir, index)

	return nil
}
