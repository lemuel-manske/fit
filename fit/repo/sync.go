package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"fit/fit/fs"
	"fit/fit/protocol"
	"fit/fit/transport"
)

func Sync(dir string, ctx context.Context) error {
	if err := fs.RequireInitialized(dir); err != nil {
		return err
	}

	config, err := fs.LoadConfig(dir)
	if err != nil {
		return err
	}

	t, err := transport.NewRabbitMQ(config.URL)
	if err != nil {
		return err
	}
	defer t.Close()

	// First resume previous incomplete syncs.
	if err = resumeIncompleteSyncs(
		dir,
		ctx,
		t,
		config.RepositoryID,
	); err != nil {
		return err
	}

	envelope, err := protocol.NewEnvelope(
		protocol.RepositoryDiscover,
		config.RepositoryID,
		config.PeerID,
		protocol.RepositoryDiscoverPayload{
			RepositoryID: config.RepositoryID,
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
		transport.DiscoveryExchange,
		transport.Message{
			Body:        body,
			ContentType: "application/json",
		},
	)
	if err != nil {
		return err
	}

	found := false

	for response := range responses {
		envelope, err := protocol.Decode(response.Body)
		if err != nil {
			continue
		}

		if envelope.Type != protocol.RepositoryOffer {
			continue
		}

		offer, err :=
			protocol.Payload[protocol.RepositoryOfferPayload](envelope)
		if err != nil {
			continue
		}

		if offer.RepositoryID != config.RepositoryID {
			continue
		}

		if offer.PeerID == config.PeerID {
			continue
		}

		found = true

		// Persist intent BEFORE downloading.
		if err = fs.MarkSyncHead(
			dir,
			offer.PeerID,
			offer.Head,
			false,
		); err != nil {
			return err
		}

		if err = FetchHead(
			dir,
			ctx,
			t,
			config.RepositoryID,
			offer.Head,
		); err != nil {
			return fmt.Errorf(
				"sync peer %s: %w",
				offer.PeerID,
				err,
			)
		}

		if err = fs.WritePeerRef(
			dir,
			offer.PeerID,
			offer.Head,
		); err != nil {
			return err
		}

		if err = fs.MarkSyncHead(
			dir,
			offer.PeerID,
			offer.Head,
			true,
		); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("no remote peers found")
	}

	return nil
}

func resumeIncompleteSyncs(
	dir string,
	ctx context.Context,
	t transport.Transport,
	repositoryID fs.RepositoryID,
) error {
	state, err := fs.LoadSyncState(dir)
	if err != nil {
		return err
	}

	for peerID, head := range state.Heads {
		if head.Complete {
			continue
		}

		if err = FetchHead(
			dir,
			ctx,
			t,
			repositoryID,
			head.CommitID,
		); err != nil {
			return fmt.Errorf(
				"resume sync peer %s: %w",
				peerID,
				err,
			)
		}

		if err = fs.WritePeerRef(
			dir,
			peerID,
			head.CommitID,
		); err != nil {
			return err
		}

		if err = fs.MarkSyncHead(
			dir,
			peerID,
			head.CommitID,
			true,
		); err != nil {
			return err
		}
	}

	return nil
}
