package repo

import (
	"context"
	"fmt"
	"os"
	"time"

	"encoding/json"

	"fit/fit/fs"
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

	discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	envelope, err := protocol.NewEnvelope(
		protocol.RepositoryDiscover,
		"",
		"",
		protocol.RepositoryDiscoverPayload{},
	) // do not apply any filters (retrieve all)
	if err != nil {
		return err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	responses, err := t.Request(
		discoveryCtx,
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

		copied := offer
		selected = &copied
	}

	if selected == nil {
		return fmt.Errorf("repository not found: %s", selector)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	config := &fs.Config{
		FormatVersion:  fs.GlobalFormatVersion,
		PeerID:         fs.PeerID(uuid.NewString()),
		RepositoryID:   selected.RepositoryID,
		RepositoryName: selected.RepositoryName,
		URL:            RemoteURL,
	}

	if err := InitWithConfig(targetDir, config); err != nil {
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

	return ForceCheckout(targetDir, selected.Head)
}
