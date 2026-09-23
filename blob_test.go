package fit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlobIDIsSHA256OfContent(t *testing.T) {
	got := BlobID([]byte("hello"))

	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	assert.Equal(t, want, string(got), "BlobID should be the SHA256 hash of the content")
}

func TestBlobIDIsDeterministic(t *testing.T) {
	content := []byte("hello")

	blobID1 := BlobID(content)
	blobID2 := BlobID(content)

	assert.Equal(t, blobID1, blobID2, "BlobID should be deterministic for the same content")
}

func TestDifferentASCIICaseProducesDifferentBlobIDs(t *testing.T) {
	content1 := []byte("a")
	content2 := []byte("A")

	blobID1 := BlobID(content1)
	blobID2 := BlobID(content2)

	assert.NotEqual(t, blobID1, blobID2, "Different ASCII case should produce different BlobIDs")
}

func TestDifferentContentProducesDifferentBlobIDs(t *testing.T) {
	content1 := []byte("hello")
	content2 := []byte("world")

	blobID1 := BlobID(content1)
	blobID2 := BlobID(content2)

	assert.NotEqual(t, blobID1, blobID2, "Different content should produce different BlobIDs")
}

func TestGetBlobRejectsContentThatDoesNotMatchItsID(t *testing.T) {
	dir := t.TempDir()

	store := NewFsBlobStore(dir)

	id, err := store.Put([]byte("hello"))
	require.NoError(t, err)

	// write different content to the file
	_, err = WriteFile(dir, "blobs", string(id), "world")
	require.NoError(t, err)

	_, err = store.Get(id)
	assert.Error(t, err, "Get should return an error if the content does not match its ID")
}

func TestPutSameBlobTwiceDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()

	store := NewFsBlobStore(dir)

	id, err := store.Put([]byte("hello"))
	require.NoError(t, err)

	// write different content to the file
	_, err = WriteFile(dir, "blobs", string(id), "world")
	require.NoError(t, err)

	// put the same blob again
	id2, err := store.Put([]byte("hello"))
	require.NoError(t, err)

	assert.Equal(t, id, id2, "Put should return the same ID for the same content")

	// get the blob and check that it is still "hello"
	content, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, []byte("hello"), content, "Get should return the original content")
}

func TestGetNonExistentBlobReturnsError(t *testing.T) {
	dir := t.TempDir()

	store := NewFsBlobStore(dir)

	_, err := store.Get("nonexistent")
	assert.Error(t, err, "Get should return an error for a non-existent blob")
}
