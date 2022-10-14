package pubsub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPubSub_GooglePubSubClient_Ack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := NewFakeGooglePubSubClient(ctx)
	require.NoError(t, err)

	want := 5
	for i := 0; i < want; i++ {
		client.Publish([]byte("test"))
	}

	client.Subscribe(func(m *PubSubMsg) {
		defer cancel()
		m.Ack()
	})

	require.Equal(t, want, client.AckCount)
}

func TestPubSub_GooglePubSubClient_Nack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := NewFakeGooglePubSubClient(ctx)
	require.NoError(t, err)

	client.Publish([]byte("test"))

	client.Subscribe(func(m *PubSubMsg) {
		defer cancel()
		m.Nack()
	})

	require.Equal(t, 1, client.NackCount)
}
