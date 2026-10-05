package cmd

import (
	"context"
	"encoding/json"

	"fit/fit/internal"
	"fit/fit/protocol"
	"fit/fit/transport"
)

func Serve(dir string, ctx context.Context) error {
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

	discoverySub, err := t.Subscribe(ctx, transport.DiscoveryExchange)
	if err != nil {
		return err
	}
	defer discoverySub.Close()

	repoSub, err := t.Subscribe(
		ctx,
		transport.RepositoryExchange(config.RepositoryID),
	)
	if err != nil {
		return err
	}
	defer repoSub.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case delivery := <-discoverySub.Messages():
			if err := handleDiscovery(dir, t, delivery, ctx); err != nil {
				if delivery.Nack != nil {
					_ = delivery.Nack(false)
				}
				continue
			}

			if delivery.Ack != nil {
				_ = delivery.Ack()
			}

		case delivery := <-repoSub.Messages():
			if err := handleRepositoryRequest(dir, t, delivery, ctx); err != nil {
				if delivery.Nack != nil {
					_ = delivery.Nack(false)
				}
				continue
			}

			if delivery.Ack != nil {
				_ = delivery.Ack()
			}
		}
	}
}

func handleDiscovery(
	dir string,
	t transport.Transport,
	d transport.Delivery,
	ctx context.Context,
) error {
	envelope, err := protocol.Decode(d.Body)
	if err != nil {
		return err
	}

	if envelope.Type != protocol.RepositoryDiscover {
		return nil
	}

	request, err := protocol.Payload[protocol.RepositoryDiscoverPayload](envelope)
	if err != nil {
		return err
	}

	config, err := internal.LoadConfig(dir)
	if err != nil {
		return err
	}

	if request.RepositoryID != "" &&
		request.RepositoryID != config.RepositoryID {
		return nil
	}

	if request.RepositoryName != "" &&
		request.RepositoryName != config.RepositoryName {
		return nil
	}

	head, err := internal.ReadHEAD(dir)
	if err != nil {
		return err
	}

	response, err := protocol.NewEnvelope(
		protocol.RepositoryOffer,
		config.RepositoryID,
		config.PeerID,
		protocol.RepositoryOfferPayload{
			RepositoryID:   config.RepositoryID,
			RepositoryName: config.RepositoryName,
			PeerID:         config.PeerID,
			Head:           head,
		},
	)
	if err != nil {
		return err
	}

	response.CorrelationID = d.CorrelationID

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return t.Reply(
		ctx,
		d.ReplyTo,
		transport.Message{
			Body:          data,
			ContentType:   "application/json",
			CorrelationID: d.CorrelationID,
		},
	)
}

func handleRepositoryRequest(
	dir string,
	t transport.Transport,
	d transport.Delivery,
	ctx context.Context,
) error {
	envelope, err := protocol.Decode(d.Body)
	if err != nil {
		return err
	}

	switch envelope.Type {
	case protocol.HeadRequest:
		return handleHeadRequest(dir, t, d, ctx)

	case protocol.CommitRequest:
		return handleCommitRequest(dir, t, d, ctx, envelope)

	case protocol.BlobRequest:
		return handleBlobRequest(dir, t, d, ctx, envelope)

	default:
		return nil
	}
}

func handleHeadRequest(
	dir string,
	t transport.Transport,
	d transport.Delivery,
	ctx context.Context,
) error {
	config, err := internal.LoadConfig(dir)
	if err != nil {
		return err
	}

	head, err := internal.ReadHEAD(dir)
	if err != nil {
		return err
	}

	envelope, err := protocol.NewEnvelope(
		protocol.HeadAnnounce,
		config.RepositoryID,
		config.PeerID,
		protocol.HeadAnnouncePayload{
			Head: head,
		},
	)
	if err != nil {
		return err
	}

	return reply(ctx, t, d, envelope)
}

func handleCommitRequest(
	dir string,
	t transport.Transport,
	d transport.Delivery,
	ctx context.Context,
	request protocol.Envelope,
) error {
	payload, err :=
		protocol.Payload[protocol.CommitRequestPayload](request)
	if err != nil {
		return err
	}

	store := internal.NewCommitStore(dir)

	commit, err := store.Get(payload.CommitID)
	if err != nil {
		return err
	}

	config, err := internal.LoadConfig(dir)
	if err != nil {
		return err
	}

	envelope, err := protocol.NewEnvelope(
		protocol.CommitResponse,
		config.RepositoryID,
		config.PeerID,
		protocol.CommitResponsePayload{
			Commit: commit,
		},
	)
	if err != nil {
		return err
	}

	return reply(ctx, t, d, envelope)
}

func handleBlobRequest(
	dir string,
	t transport.Transport,
	d transport.Delivery,
	ctx context.Context,
	request protocol.Envelope,
) error {
	payload, err :=
		protocol.Payload[protocol.BlobRequestPayload](request)
	if err != nil {
		return err
	}

	store := internal.NewBlobStore(dir)

	data, err := store.Get(payload.Hash)
	if err != nil {
		return err
	}

	config, err := internal.LoadConfig(dir)
	if err != nil {
		return err
	}

	envelope, err := protocol.NewEnvelope(
		protocol.BlobResponse,
		config.RepositoryID,
		config.PeerID,
		protocol.BlobResponsePayload{
			Hash: payload.Hash,
			Data: data,
		},
	)
	if err != nil {
		return err
	}

	return reply(ctx, t, d, envelope)
}

func reply(
	ctx context.Context,
	t transport.Transport,
	d transport.Delivery,
	envelope protocol.Envelope,
) error {
	envelope.CorrelationID = d.CorrelationID

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return t.Reply(
		ctx,
		d.ReplyTo,
		transport.Message{
			Body:          data,
			ContentType:   "application/json",
			CorrelationID: d.CorrelationID,
		},
	)
}
