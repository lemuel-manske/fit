package repo

import (
	"context"
	"encoding/json"

	"fit/fit/protocol"
	"fit/fit/transport"
)

func DiscoverAll(
	dir string,
	ctx context.Context,
) (<-chan protocol.RepositoryOfferPayload, error) {
	t, err := transport.NewRabbitMQ()
	if err != nil {
		return nil, err
	}

	envelope, err := protocol.NewEnvelope(
		protocol.RepositoryDiscover,
		"",
		"",
		protocol.RepositoryDiscoverPayload{},
	)
	if err != nil {
		_ = t.Close()
		return nil, err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		_ = t.Close()
		return nil, err
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
		_ = t.Close()
		return nil, err
	}

	out := make(chan protocol.RepositoryOfferPayload)

	go func() {
		defer close(out)
		defer t.Close()

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

			select {
			case out <- offer:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
