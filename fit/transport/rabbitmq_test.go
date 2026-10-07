package transport

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRabbitMQRequiresURL(t *testing.T) {
	t.Setenv("FIT_BROKER_URL", "")
	_, err := NewRabbitMQ()
	require.ErrorContains(t, err, "FIT_BROKER_URL")
}

func TestRabbitMQRetryBinaryReplyAndDeadLetters(t *testing.T) {
	if os.Getenv("FIT_BROKER_URL") == "" {
		t.Skip("FIT_BROKER_URL is required for integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := NewRabbitMQ()
	require.NoError(t, err)
	defer client.Close()
	server, err := NewRabbitMQ()
	require.NoError(t, err)
	defer server.Close()
	exchange := "fit.test." + uuid.NewString()
	sub, err := server.Subscribe(ctx, exchange)
	require.NoError(t, err)
	defer sub.Close()

	responses, err := client.Request(ctx, exchange, Message{Body: []byte("request")})
	require.NoError(t, err)
	var first Delivery
	select {
	case first = <-sub.Messages():
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	require.NoError(t, first.Ack()) // Simulate a lost response.
	var retry Delivery
	select {
	case retry = <-sub.Messages():
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	require.Equal(t, first.CorrelationID, retry.CorrelationID)
	require.Equal(t, first.ReplyTo, retry.ReplyTo)
	data := []byte{0, 255, 128, 10}
	require.NoError(t, server.Reply(ctx, retry.ReplyTo, Message{
		Body: data,
		ContentType: "application/octet-stream",
		CorrelationID: retry.CorrelationID,
		Headers: map[string]string{"blobHash": "hash"},
	}))
	require.NoError(t, retry.Ack())
	select {
	case response := <-responses:
		require.Equal(t, data, response.Body)
		require.Equal(t, "hash", response.Headers["blobHash"])
		require.Equal(t, "application/octet-stream", response.ContentType)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	poison := []byte(uuid.NewString())
	require.NoError(t, client.Publish(ctx, exchange, Message{Body: poison}))
	select {
	case delivery := <-sub.Messages():
		require.Equal(t, poison, delivery.Body)
		require.NoError(t, delivery.Nack(false))
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	ch, err := server.(*RabbitMQ).conn.Channel()
	require.NoError(t, err)
	defer ch.Close()
	for {
		dead, ok, err := ch.Get("fit.dlq", true)
		require.NoError(t, err)
		if ok && string(dead.Body) == string(poison) {
			require.Contains(t, dead.Headers, "x-death")
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("message did not reach DLQ")
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func TestRabbitMQRetryLimit(t *testing.T) {
	if os.Getenv("FIT_BROKER_URL") == "" {
		t.Skip("FIT_BROKER_URL is required for integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := NewRabbitMQ()
	require.NoError(t, err)
	defer client.Close()
	exchange := "fit.test." + uuid.NewString()
	sub, err := client.Subscribe(ctx, exchange)
	require.NoError(t, err)
	defer sub.Close()
	responses, err := client.Request(ctx, exchange, Message{Body: []byte("unanswered")})
	require.NoError(t, err)
	count := 0
	for {
		select {
		case delivery := <-sub.Messages():
			count++
			require.NoError(t, delivery.Ack())
		case _, ok := <-responses:
			require.False(t, ok)
			require.Equal(t, 3, count)
			return
		case <-ctx.Done():
			t.Fatal("request did not exhaust retries")
		}
	}
}

func TestRabbitMQPermissions(t *testing.T) {
	if os.Getenv("FIT_BROKER_URL") == "" {
		t.Skip("FIT_BROKER_URL is required for integration tests")
	}
	client, err := NewRabbitMQ()
	require.NoError(t, err)
	defer client.Close()
	err = client.Publish(context.Background(), "unrelated.exchange", Message{})
	require.Error(t, err)
}
