package transport

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQ(url string) (Transport, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn: conn,
		ch:   ch,
	}, nil
}

func (r *RabbitMQ) Publish(
	ctx context.Context,
	exchange string,
	message Message,
) error {
	if err := r.ensureExchange(exchange); err != nil {
		return err
	}

	headers := amqp.Table{}
	for key, value := range message.Headers {
		headers[key] = value
	}

	return r.ch.PublishWithContext(
		ctx,
		exchange,
		"",    // fanout exchange does not use routing keys
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:   message.ContentType,
			Body:          message.Body,
			ReplyTo:       message.ReplyTo,
			CorrelationId: message.CorrelationID,
			Headers:       headers,
		},
	)
}

func (r *RabbitMQ) Subscribe(
	ctx context.Context,
	exchange string,
) (Subscription, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	if err = ch.ExchangeDeclare(
		exchange,
		"fanout",
		false, // durable
		true,  // autoDelete
		false, // internal
		false, // noWait
		nil,
	); err != nil {
		_ = ch.Close()
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		"",    // let RabbitMQ generate a unique queue name
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, err
	}

	if err = ch.QueueBind(
		queue.Name,
		"", // let RabbitMQ handle binding name
		exchange,
		false, // noWait
		nil,
	); err != nil {
		_ = ch.Close()
		return nil, err
	}

	deliveries, err := ch.Consume(
		queue.Name,
		"",    // let RabbitMQ generate a unique consumer tag
		false, // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		_ = ch.Close()

		return nil, err
	}

	out := make(chan Delivery)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return

			case d, ok := <-deliveries:
				if !ok {
					return
				}

				headers := map[string]string{}

				for key, value := range d.Headers {
					headers[key] = fmt.Sprint(value)
				}

				out <- Delivery{
					Message: Message{
						Body:          d.Body,
						ContentType:   d.ContentType,
						ReplyTo:       d.ReplyTo,
						CorrelationID: d.CorrelationId,
						Headers:       headers,
					},
					Ack: func() error {
						return d.Ack(false)
					},
					Nack: func(requeue bool) error {
						return d.Nack(false, requeue)
					},
				}
			}
		}
	}()

	return &RabbitMQSubscription{
		ch:    ch,
		queue: queue.Name,
		msgs:  out,
	}, nil
}

func (r *RabbitMQ) Reply(
	ctx context.Context,
	queue string,
	message Message,
) error {
	return r.ch.PublishWithContext(
		ctx,
		"",
		queue,
		false, // immediate
		false, // mandatory
		amqp.Publishing{
			ContentType:   message.ContentType,
			Body:          message.Body,
			CorrelationId: message.CorrelationID,
		},
	)
}

func (r *RabbitMQ) Request(
	ctx context.Context,
	exchange string,
	message Message,
) (<-chan Message, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		"",
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, err
	}

	correlationID := uuid.NewString()

	message.ReplyTo = queue.Name
	message.CorrelationID = correlationID

	deliveries, err := ch.Consume(
		queue.Name,
		"",
		true,  // autoAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, err
	}

	if err := r.Publish(ctx, exchange, message); err != nil {
		_ = ch.Close()
		return nil, err
	}

	out := make(chan Message)

	go func() {
		defer close(out)
		defer ch.Close()

		for {
			select {
			case <-ctx.Done():
				return

			case d, ok := <-deliveries:
				if !ok {
					return
				}

				if d.CorrelationId != correlationID {
					continue
				}

				out <- Message{
					Body:          d.Body,
					ContentType:   d.ContentType,
					ReplyTo:       d.ReplyTo,
					CorrelationID: d.CorrelationId,
				}
			}
		}
	}()

	return out, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.ch.Close(); err != nil {
		return err
	}

	return r.conn.Close()
}

func (r *RabbitMQ) ensureExchange(name string) error {
	return r.ch.ExchangeDeclare(
		name,
		"fanout",
		false, // durable
		true,  // autoDelete
		false, // internal
		false, // noWait
		nil,   // no args
	)
}

type RabbitMQSubscription struct {
	ch    *amqp.Channel
	queue string
	msgs  chan Delivery
}

func (s *RabbitMQSubscription) Messages() <-chan Delivery {
	return s.msgs
}

func (s *RabbitMQSubscription) Close() error {
	return s.ch.Close()
}
