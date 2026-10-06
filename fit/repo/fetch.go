package repo

import (
	"context"
	"errors"
	"fmt"
	"os"

	"encoding/json"

	"fit/fit/fs"
	"fit/fit/protocol"
	"fit/fit/transport"
)

func FetchHead(
	dir string,
	ctx context.Context,
	t transport.Transport,
	repositoryID fs.RepositoryID,
	head fs.CommitID,
) error {
	if head == "" {
		return nil
	}

	return fetchCommit(
		dir,
		ctx,
		t,
		repositoryID,
		head,
		map[fs.CommitID]bool{},
	)
}

func fetchCommit(
	dir string,
	ctx context.Context,
	t transport.Transport,
	repositoryID fs.RepositoryID,
	id fs.CommitID,
	visited map[fs.CommitID]bool,
) error {
	if visited[id] {
		return nil
	}

	visited[id] = true

	store := fs.NewCommitStore(dir)

	commit, err := store.Get(id)

	if err == nil {
		return fetchCommitDependencies(
			dir,
			ctx,
			t,
			repositoryID,
			commit,
			visited,
		)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	commit, err = requestCommit(
		ctx,
		t,
		repositoryID,
		id,
	)
	if err != nil {
		return err
	}

	if commit.RepositoryID != repositoryID {
		return fmt.Errorf("commit belongs to another repository")
	}

	storedID, err := store.Put(commit)
	if err != nil {
		return err
	}

	if storedID != id {
		return fmt.Errorf(
			"invalid commit: expected %s, got %s",
			id,
			storedID,
		)
	}

	return fetchCommitDependencies(
		dir,
		ctx,
		t,
		repositoryID,
		commit,
		visited,
	)
}

func fetchCommitDependencies(
	dir string,
	ctx context.Context,
	t transport.Transport,
	repositoryID fs.RepositoryID,
	commit fs.Commit,
	visited map[fs.CommitID]bool,
) error {
	for _, parent := range commit.Parents {
		if err := fetchCommit(
			dir,
			ctx,
			t,
			repositoryID,
			parent,
			visited,
		); err != nil {
			return err
		}
	}

	for _, hash := range commit.Files {
		if err := fetchBlob(
			dir,
			ctx,
			t,
			repositoryID,
			hash,
		); err != nil {
			return err
		}
	}

	return nil
}

func requestCommit(
	ctx context.Context,
	t transport.Transport,
	repositoryID fs.RepositoryID,
	id fs.CommitID,
) (fs.Commit, error) {
	envelope, err := protocol.NewEnvelope(
		protocol.CommitRequest,
		repositoryID,
		"",
		protocol.CommitRequestPayload{
			CommitID: id,
		},
	)
	if err != nil {
		return fs.Commit{}, err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fs.Commit{}, err
	}

	responses, err := t.Request(
		ctx,
		transport.RepositoryExchange(repositoryID),
		transport.Message{
			Body:        body,
			ContentType: "application/json",
		},
	)
	if err != nil {
		return fs.Commit{}, err
	}

	for response := range responses {
		envelope, err := protocol.Decode(response.Body)
		if err != nil {
			continue
		}

		if envelope.Type != protocol.CommitResponse {
			continue
		}

		payload, err :=
			protocol.Payload[protocol.CommitResponsePayload](envelope)
		if err != nil {
			continue
		}

		if payload.Commit.ID != id {
			continue
		}

		return payload.Commit, nil
	}

	return fs.Commit{}, fmt.Errorf("commit not found: %s", id)
}

func fetchBlob(
	dir string,
	ctx context.Context,
	t transport.Transport,
	repositoryID fs.RepositoryID,
	hash fs.Hash,
) error {
	store := fs.NewBlobStore(dir)

	if _, err := store.Get(hash); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	envelope, err := protocol.NewEnvelope(
		protocol.BlobRequest,
		repositoryID,
		"",
		protocol.BlobRequestPayload{
			Hash: hash,
		},
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	responses, err := t.Request(
		ctx,
		transport.RepositoryExchange(repositoryID),
		transport.Message{
			Body:        body,
			ContentType: "application/json",
		},
	)
	if err != nil {
		return err
	}

	for response := range responses {
		envelope, err := protocol.Decode(response.Body)
		if err != nil {
			continue
		}

		if envelope.Type != protocol.BlobResponse {
			continue
		}

		payload, err :=
			protocol.Payload[protocol.BlobResponsePayload](envelope)
		if err != nil {
			continue
		}

		if payload.Hash != hash {
			continue
		}

		storedHash, err := store.Put(payload.Data)
		if err != nil {
			return err
		}

		if storedHash != hash {
			return fmt.Errorf(
				"invalid blob: expected %s, got %s",
				hash,
				storedHash,
			)
		}

		return nil
	}

	return fmt.Errorf("blob not found: %s", hash)
}
