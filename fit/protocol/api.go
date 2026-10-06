// Package protocol defines the message types and payloads used for communication between peers in the Fit "protocol".
package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"fit/fit/fs"

	"github.com/google/uuid"
)

type MessageType string

const (
	GlobalProtocolVersion = 1

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
	ProtocolVersion int             `json:"protocolVersion"`
	MessageID       string          `json:"messageId"`
	Type            MessageType     `json:"type"`
	RepositoryID    fs.RepositoryID `json:"repositoryId,omitempty"`
	SenderPeerID    fs.PeerID       `json:"senderPeerId"`
	CorrelationID   string          `json:"correlationId,omitempty"`
	SentAt          time.Time       `json:"sentAt"`
	Payload         json.RawMessage `json:"payload,omitempty"`
}

func NewEnvelope(
	messageType MessageType,
	repositoryID fs.RepositoryID,
	peerID fs.PeerID,
	payload any,
) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}

	return Envelope{
		ProtocolVersion: GlobalProtocolVersion,
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

	if envelope.ProtocolVersion != GlobalProtocolVersion {
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

// Payload decodes the payload of an envelope into the specified type T.
func Payload[T any](envelope Envelope) (T, error) {
	var value T

	if err := json.Unmarshal(envelope.Payload, &value); err != nil {
		return value, err
	}

	return value, nil
}

type RepositoryDiscoverPayload struct {
	RepositoryID   fs.RepositoryID `json:"repositoryId,omitempty"`
	RepositoryName string          `json:"repositoryName,omitempty"`
}

type RepositoryOfferPayload struct {
	RepositoryID   fs.RepositoryID `json:"repositoryId"`
	RepositoryName string          `json:"repositoryName"`
	PeerID         fs.PeerID       `json:"peerId"`
	Head           fs.CommitID     `json:"head"`
}

type HeadRequestPayload struct{}

type HeadAnnouncePayload struct {
	Head fs.CommitID `json:"head"`
}

type CommitRequestPayload struct {
	CommitID fs.CommitID `json:"commitId"`
}

type CommitResponsePayload struct {
	Commit fs.Commit `json:"commit"`
}

type BlobRequestPayload struct {
	Hash fs.Hash `json:"hash"`
}

type BlobResponsePayload struct {
	Hash fs.Hash `json:"hash"`
	Data []byte  `json:"data"`
}
