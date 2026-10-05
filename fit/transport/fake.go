package transport

import (
	"context"
	"sync"
)

type FakeTransport struct {
	mu          sync.Mutex
	Subscribers map[string][]chan Delivery
}

func NewFakeTransport() *FakeTransport {
	return &FakeTransport{
		Subscribers: make(map[string][]chan Delivery),
	}
}

func (f *FakeTransport) Publish(
	ctx context.Context,
	exchange string,
	message Message,
) error {
	f.mu.Lock()
	subs := append([]chan Delivery(nil), f.Subscribers[exchange]...)
	f.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- Delivery{Message: message}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

type fakeSubscription struct {
	ch chan Delivery
}

func (s *fakeSubscription) Messages() <-chan Delivery {
	return s.ch
}

func (s *fakeSubscription) Close() error {
	return nil
}

func (f *FakeTransport) Subscribe(
	ctx context.Context,
	exchange string,
) (Subscription, error) {
	ch := make(chan Delivery, 16)

	f.mu.Lock()
	f.Subscribers[exchange] = append(f.Subscribers[exchange], ch)
	f.mu.Unlock()

	return &fakeSubscription{ch: ch}, nil
}

func (f *FakeTransport) Reply(
	ctx context.Context,
	queue string,
	message Message,
) error {
	return f.Publish(ctx, queue, message)
}

func (f *FakeTransport) Request(
	ctx context.Context,
	exchange string,
	message Message,
) (<-chan Message, error) {
	responses := make(chan Message, 16)

	message.ReplyTo = "fake.reply"

	sub, err := f.Subscribe(ctx, message.ReplyTo)
	if err != nil {
		return nil, err
	}

	if err := f.Publish(ctx, exchange, message); err != nil {
		return nil, err
	}

	go func() {
		defer close(responses)
		defer sub.Close()

		for {
			select {
			case delivery := <-sub.Messages():
				responses <- delivery.Message

			case <-ctx.Done():
				return
			}
		}
	}()

	return responses, nil
}

func (f *FakeTransport) Close() error {
	return nil
}
