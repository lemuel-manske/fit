// Package transport defines the API for message transport.
package transport

import "context"

type Message struct {
	Body          []byte
	ContentType   string
	ReplyTo       string
	CorrelationID string
	Headers       map[string]string
}

type Delivery struct {
	Message
	Ack  func() error
	Nack func(requeue bool) error
}

type Subscription interface {
	Messages() <-chan Delivery
	Close() error
}

type Transport interface {
	Publish(
		ctx context.Context,
		exchange string,
		message Message,
	) error

	Reply(
		ctx context.Context,
		queue string,
		message Message,
	) error

	Subscribe(
		ctx context.Context,
		exchange string,
	) (Subscription, error)

	Request(
		ctx context.Context,
		exchange string,
		message Message,
	) (<-chan Message, error)

	Close() error
}
