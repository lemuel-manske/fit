package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	"fit/fit/fs"
	"fit/fit/protocol"
	"fit/fit/transport"

	"github.com/google/uuid"
)

func Serve(dir string, ctx context.Context) error {
	if err := fs.RequireInitialized(dir); err != nil {
		return err
	}

	lock, err := fs.Lock(dir, "serve", ctx, false)
	if err != nil {
		return err
	}
	defer lock.Close()

	backoff := time.Second

	for {
		err := serveOnce(dir, ctx)

		if ctx.Err() != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil

		case <-time.After(backoff):
		}

		if err == nil {
			backoff = time.Second
			continue
		}

		backoff *= 2

		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
	}
}

func serveOnce(dir string, ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	config, err := fs.LoadConfig(dir)
	if err != nil {
		return err
	}

	t, err := transport.NewRabbitMQ()
	if err != nil {
		return err
	}
	defer t.Close()

	discoverySub, err := t.Subscribe(
		ctx,
		transport.DiscoveryExchange,
	)
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

	// every successful connection/reconnection publishes current HEAD.
	if err = announceHead(dir, ctx, t); err != nil {
		return err
	}

	go watchHead(dir, ctx, t)

	for {
		select {
		case <-ctx.Done():
			return nil

		case delivery, ok := <-discoverySub.Messages():
			if !ok {
				return fmt.Errorf("discovery subscription closed")
			}

			if err = handleDiscovery(
				dir,
				t,
				delivery,
				ctx,
			); err != nil {
				if delivery.Nack != nil {
					_ = delivery.Nack(false)
				}
				continue
			}

			if delivery.Ack != nil {
				_ = delivery.Ack()
			}

		case delivery, ok := <-repoSub.Messages():
			if !ok {
				return fmt.Errorf("repository subscription closed")
			}

			if err = handleRepositoryRequest(
				dir,
				t,
				delivery,
				ctx,
			); err != nil {
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

	config, err := fs.LoadConfig(dir)
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

	head, err := fs.ReadHEAD(dir)
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

	case protocol.HeadAnnounce:
		return handleHeadAnnounce(dir, envelope, ctx)

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
	config, err := fs.LoadConfig(dir)
	if err != nil {
		return err
	}

	head, err := fs.ReadHEAD(dir)
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

	store := fs.NewCommitStore(dir)

	commit, err := store.Get(payload.CommitID)
	if err != nil {
		return err
	}

	config, err := fs.LoadConfig(dir)
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

	store := fs.NewBlobStore(dir)

	data, err := store.Get(payload.Hash)
	if err != nil {
		return err
	}

	config, err := fs.LoadConfig(dir)
	if err != nil {
		return err
	}

	return t.Reply(
		ctx,
		d.ReplyTo,
		transport.Message{
			Body:          data,
			ContentType:   "application/octet-stream",
			CorrelationID: d.CorrelationID,
			Headers: map[string]string{
				"protocolVersion": fmt.Sprint(protocol.GlobalProtocolVersion),
				"messageId":        uuid.NewString(),
				"type":             string(protocol.BlobResponse),
				"repositoryId":     string(config.RepositoryID),
				"senderPeerId":     string(config.PeerID),
				"sentAt":           time.Now().UTC().Format(time.RFC3339Nano),
				"blobHash":         string(payload.Hash),
			},
		},
	)
}

func handleHeadAnnounce(
	dir string,
	envelope protocol.Envelope,
	ctx context.Context,
) error {
	lock, err := fs.Lock(dir, "write", ctx, false)
	// Announcements are advisory; discovery during sync recovers a skipped head.
	// Never block request handling behind a CLI operation waiting for replies.
	if errors.Is(err, fs.ErrLockBusy) {
		return nil
	}
	if err != nil {
		return err
	}
	defer lock.Close()

	config, err := fs.LoadConfig(dir)
	if err != nil {
		return err
	}

	if envelope.RepositoryID != config.RepositoryID {
		return nil
	}

	if envelope.SenderPeerID == config.PeerID {
		return nil
	}

	payload, err :=
		protocol.Payload[protocol.HeadAnnouncePayload](envelope)
	if err != nil {
		return err
	}

	return fs.WritePeerRef(
		dir,
		envelope.SenderPeerID,
		payload.Head,
	)
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

func watchHead(
	dir string,
	ctx context.Context,
	t transport.Transport,
) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var last fs.CommitID

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			head, err := fs.ReadHEAD(dir)
			if err != nil {
				continue
			}

			if head == last {
				continue
			}

			if err := announceHead(dir, ctx, t); err == nil {
				last = head
			}
		}
	}
}

func announceHead(
	dir string,
	ctx context.Context,
	t transport.Transport,
) error {
	config, err := fs.LoadConfig(dir)
	if err != nil {
		return err
	}

	head, err := fs.ReadHEAD(dir)
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

	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return t.Publish(
		ctx,
		transport.RepositoryExchange(config.RepositoryID),
		transport.Message{
			Body:        body,
			ContentType: "application/json",
		},
	)
}
