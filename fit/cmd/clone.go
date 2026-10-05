package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"fit/fit/internal"
	"fit/fit/protocol"
	"fit/fit/transport"

	"github.com/google/uuid"
)

func Clone(
	ctx context.Context,
	selector string,
	targetDir string,
) error {
	t, err := transport.NewRabbitMQ(RemoteURL)
	if err != nil {
		return err
	}
	defer t.Close()

	envelope, err := protocol.NewEnvelope(
		protocol.RepositoryDiscover,
		"",
		"",
		protocol.RepositoryDiscoverPayload{},
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

	var selected *protocol.RepositoryOfferPayload

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

		if string(offer.RepositoryID) != selector &&
			offer.RepositoryName != selector {
			continue
		}

		if selected != nil &&
			selected.RepositoryID != offer.RepositoryID {
			return fmt.Errorf(
				"repository selector %q is ambiguous",
				selector,
			)
		}

		copy := offer
		selected = &copy
	}

	if selected == nil {
		return fmt.Errorf("repository not found: %s", selector)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	config := &internal.Config{
		FormatVersion:  internal.FormatVersion,
		PeerID:         internal.PeerID(uuid.NewString()),
		RepositoryID:   selected.RepositoryID,
		RepositoryName: selected.RepositoryName,
		URL:            RemoteURL,
	}

	if err := InitWith(
		targetDir,
		config,
	); err != nil {
		return err
	}

	if err := FetchHead(
		targetDir,
		ctx,
		t,
		selected.RepositoryID,
		selected.Head,
	); err != nil {
		return err
	}

	if err := internal.WriteHEAD(
		targetDir,
		selected.Head,
	); err != nil {
		return err
	}

	return Checkout(targetDir, selected.Head)
}
