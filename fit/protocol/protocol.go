package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"fit/fit/internal"

	"github.com/google/uuid"
)

const Version = 1

type MessageType string

const (
	RepositoryDiscover MessageType = "repository.discover"
	RepositoryOffer    MessageType = "repository.offer"

	HeadRequest  MessageType = "head.request"
	HeadAnnounce MessageType = "head.announce"

	CommitRequest  MessageType = "commit.request"
	CommitResponse MessageType = "commit.response"

	BlobRequest  MessageType = "blob.request"
	BlobResponse MessageType = "blob.response"
)

type Envelope struct {
	ProtocolVersion int                   `json:"protocolVersion"`
	MessageID       string                `json:"messageId"`
	Type            MessageType           `json:"type"`
	RepositoryID    internal.RepositoryID `json:"repositoryId,omitempty"`
	SenderPeerID    internal.PeerID       `json:"senderPeerId"`
	CorrelationID   string                `json:"correlationId,omitempty"`
	SentAt          time.Time             `json:"sentAt"`
	Payload         json.RawMessage       `json:"payload,omitempty"`
}

func NewEnvelope(
	messageType MessageType,
	repositoryID internal.RepositoryID,
	peerID internal.PeerID,
	payload any,
) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}

	return Envelope{
		ProtocolVersion: Version,
		MessageID:       uuid.NewString(),
		Type:            messageType,
		RepositoryID:    repositoryID,
		SenderPeerID:    peerID,
		SentAt:          time.Now().UTC(),
		Payload:         data,
	}, nil
}

func Decode(data []byte) (Envelope, error) {
	var envelope Envelope

	if err := json.Unmarshal(data, &envelope); err != nil {
		return Envelope{}, err
	}

	if envelope.ProtocolVersion != Version {
		return Envelope{}, fmt.Errorf(
			"unsupported protocol version: %d",
			envelope.ProtocolVersion,
		)
	}

	if envelope.MessageID == "" {
		return Envelope{}, fmt.Errorf("message id is required")
	}

	return envelope, nil
}

func Payload[T any](envelope Envelope) (T, error) {
	var value T

	if err := json.Unmarshal(envelope.Payload, &value); err != nil {
		return value, err
	}

	return value, nil
}
