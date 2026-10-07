package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func loadURL() string {
	url := ""

	if envURL := os.Getenv("FIT_BROKER_URL"); envURL != "" {
		url = envURL
	}

	return url
}

func NewRabbitMQ() (Transport, error) {
	url := loadURL()

	if url == "" {
		return nil, fmt.Errorf("FIT_BROKER_URL is required")
	}

	var tlsConfig *tls.Config
	if caFile := os.Getenv("FIT_BROKER_CA_FILE"); caFile != "" {
		data, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		if !roots.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("invalid broker CA certificate")
		}
		tlsConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	}

	conn, err := amqp.DialConfig(url, amqp.Config{TLSClientConfig: tlsConfig})
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

	if err = ensureDeadLetters(ch); err != nil {
		_ = ch.Close()
		return nil, err
	}

	if err = ch.Qos(16, 0, false); err != nil {
		_ = ch.Close()
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		"",    // let RabbitMQ generate a unique queue name
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		amqp.Table{"x-dead-letter-exchange": "fit.dlx"},
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

				delivery := Delivery{
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
				select {
				case out <- delivery:
				case <-ctx.Done():
					return
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
			Headers:       messageHeaders(message),
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

	if err = r.Publish(ctx, exchange, message); err != nil {
		_ = ch.Close()
		return nil, err
	}

	out := make(chan Message)

	go func() {
		defer close(out)
		defer ch.Close()

		backoff := time.Second
		timer := time.NewTimer(backoff)
		defer timer.Stop()
		attempts := 1

		for {
			select {
			case <-ctx.Done():
				return

			case <-timer.C:
				if attempts == 3 {
					return
				}
				if err := r.Publish(ctx, exchange, message); err != nil {
					return
				}
				attempts++
				backoff *= 2
				timer.Reset(backoff)

			case d, ok := <-deliveries:
				if !ok {
					return
				}

				if d.CorrelationId != correlationID {
					continue
				}

				response := Message{
					Body:          d.Body,
					ContentType:   d.ContentType,
					ReplyTo:       d.ReplyTo,
					CorrelationID: d.CorrelationId,
				}
				response.Headers = map[string]string{}
				for key, value := range d.Headers {
					response.Headers[key] = fmt.Sprint(value)
				}
				select {
				case out <- response:
				case <-ctx.Done():
					return
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

func messageHeaders(message Message) amqp.Table {
	headers := amqp.Table{}
	for key, value := range message.Headers {
		headers[key] = value
	}
	return headers
}

func ensureDeadLetters(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare("fit.dlx", "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	queue, err := ch.QueueDeclare("fit.dlq", true, false, false, false, amqp.Table{
		"x-message-ttl": int32(86400000),
		"x-max-length":  int32(10000),
	})
	if err != nil {
		return err
	}
	return ch.QueueBind(queue.Name, "", "fit.dlx", false, nil)
}
