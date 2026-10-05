package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"fit/fit/internal"
	"fit/fit/protocol"
	"fit/fit/transport"
)

func Sync(dir string, ctx context.Context) error {
	if err := internal.RequireInitialized(dir); err != nil {
		return err
	}

	config, err := internal.LoadConfig(dir)
	if err != nil {
		return err
	}

	t, err := transport.NewRabbitMQ(config.URL)
	if err != nil {
		return err
	}
	defer t.Close()

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

		if err := FetchHead(
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

		if err := internal.WritePeerRef(
			dir,
			offer.PeerID,
			offer.Head,
		); err != nil {
			return err
		}
	}

	if !found {
		return fmt.Errorf("no remote peers found")
	}

	return nil
}
