package fit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlobIDIsSHA256OfContent(t *testing.T) {
	got := NewBlobID([]byte("hello"))

	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	require.Equal(t, want, string(got), "BlobID should be the SHA256 hash of the content")
}

func TestBlobIDIsDeterministic(t *testing.T) {
	content := []byte("hello")

	blobID1 := NewBlobID(content)
	blobID2 := NewBlobID(content)

	require.Equal(t, blobID1, blobID2, "BlobID should be deterministic for the same content")
}

func TestDifferentASCIICaseProducesDifferentBlobIDs(t *testing.T) {
	content1 := []byte("a")
	content2 := []byte("A")

	blobID1 := NewBlobID(content1)
	blobID2 := NewBlobID(content2)

	require.NotEqual(t, blobID1, blobID2, "Different ASCII case should produce different BlobIDs")
}

func TestDifferentContentProducesDifferentBlobIDs(t *testing.T) {
	content1 := []byte("hello")
	content2 := []byte("world")

	blobID1 := NewBlobID(content1)
	blobID2 := NewBlobID(content2)

	require.NotEqual(t, blobID1, blobID2, "Different content should produce different BlobIDs")
}

func TestGetBlobRejectsContentThatDoesNotMatchItsID(t *testing.T) {
	dir := t.TempDir()

	InitBlobStore(dir)

	store := NewBlobStore(dir)

	id, err := store.Put([]byte("hello"))
	require.NoError(t, err)

	err = CorruptBlob(dir, id, "world")
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err, "Get should return an error if the content does not match its ID")
}

func TestPutSameBlobTwiceDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()

	InitBlobStore(dir)

	store := NewBlobStore(dir)

	initialContent := []byte("hello")

	id, err := store.Put(initialContent)
	require.NoError(t, err)

	// write different content to the file
	err = CorruptBlob(dir, id, "world")
	require.NoError(t, err)

	// put the same blob again
	id2, err := store.Put(initialContent)
	require.NoError(t, err)

	require.Equal(t, id, id2, "Put should return the same ID for the same content")

	// get the blob and check that it is still "hello"
	content, err := store.Get(id)
	require.NoError(t, err)

	require.Equal(t, initialContent, content, "Get should return the original content")
}

func TestGetNonExistentBlobReturnsError(t *testing.T) {
	dir := t.TempDir()

	InitBlobStore(dir)

	store := NewBlobStore(dir)

	_, err := store.Get("nonexistent")
	require.Error(t, err, "Get should return an error for a non-existent blob")
}

func TestStagedBlobReturnsErrorForStagedDeletion(t *testing.T) {
	dir := t.TempDir()

	entries := make(map[string]IndexEntry)
	entries["a.txt"] = IndexEntry{
		Delete: true,
	}

	index := &Index{
		Entries: entries,
	}

	err := WriteIndex(dir, index)
	require.NoError(t, err)

	_, err = StagedBlob(dir, "a.txt")

	require.Error(t, err, "StagedBlob should return an error for a staged deletion")
}
