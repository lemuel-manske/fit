package repo

import (
	"context"
	"fmt"
	"slices"
	"time"

	"encoding/json"

	"fit/fit/fs"
	"fit/fit/protocol"
	"fit/fit/transport"
)

type SyncResult struct {
	Remotes []RemoteStatus
}

func Sync(dir string, ctx context.Context) (*SyncResult, error) {
	if err := fs.RequireInitialized(dir); err != nil {
		return nil, err
	}

	config, err := fs.LoadConfig(dir)
	if err != nil {
		return nil, err
	}

	t, err := transport.NewRabbitMQ(RemoteURL)
	if err != nil {
		return nil, err
	}
	defer t.Close()

	if err = resumeIncompleteSyncs(
		dir,
		ctx,
		t,
		config.RepositoryID,
	); err != nil {
		return nil, err
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
		return nil, err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	responses, err := t.Request(
		discoveryCtx,
		transport.DiscoveryExchange,
		transport.Message{
			Body:        body,
			ContentType: "application/json",
		},
	)
	if err != nil {
		return nil, err
	}

	headID, err := fs.ReadHEAD(dir)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Remotes: []RemoteStatus{},
	}

	seen := map[fs.PeerID]bool{}

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

		if offer.PeerID == config.PeerID || seen[offer.PeerID] {
			continue
		}

		seen[offer.PeerID] = true

		if err = fs.MarkSyncHead(
			dir,
			offer.PeerID,
			offer.Head,
			false,
		); err != nil {
			return nil, err
		}

		if err = FetchHead(
			dir,
			ctx,
			t,
			config.RepositoryID,
			offer.Head,
		); err != nil {
			return nil, fmt.Errorf(
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
			return nil, err
		}

		if err = fs.MarkSyncHead(
			dir,
			offer.PeerID,
			offer.Head,
			true,
		); err != nil {
			return nil, err
		}

		relation, err := commitRelation(dir, headID, offer.Head)
		if err != nil {
			return nil, err
		}

		result.Remotes = append(result.Remotes, RemoteStatus{
			PeerID:   offer.PeerID,
			Head:     offer.Head,
			Relation: relation,
		})
	}

	if len(result.Remotes) == 0 {
		return nil, fmt.Errorf("no remote peers found")
	}

	slices.SortFunc(result.Remotes, func(a, b RemoteStatus) int {
		if a.PeerID < b.PeerID {
			return -1
		}
		if a.PeerID > b.PeerID {
			return 1
		}
		return 0
	})

	return result, nil
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
