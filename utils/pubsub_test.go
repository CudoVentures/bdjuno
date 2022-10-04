package utils

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

func TestPubSub_GooglePubSubClient(t *testing.T) {
	projectID, topicID, subID := "p", "t", "s"
	ctx, cancel := context.WithCancel(context.Background())
	client, err := NewFakeGooglePubSubClient(ctx, projectID, subID, topicID)
	require.NoError(t, err)

	want := 5
	for i := 0; i < want; i++ {
		client.Publish([]byte("test"))
	}

	count := 0
	client.Subscribe(func(m *Message) {
		defer cancel()
		count++
		m.Ack()
	})

	require.Equal(t, want, count)
}

// todo accept ReceiveSettings as param, cuz here we need synchronous, but in other tests we would want original settings or sth
func NewFakeGooglePubSubClient(ctx context.Context, projectID string, subID string, topicID string) (*GooglePubSubClient, error) {
	server := pstest.NewServer()
	client, err := pubsub.NewClient(
		ctx,
		projectID,
		option.WithEndpoint(server.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
	)
	if err != nil {
		return nil, err
	}

	topic, err := client.CreateTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	sub, err := client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		return nil, err
	}
	sub.ReceiveSettings.Synchronous = true

	return &GooglePubSubClient{ctx: ctx, sub: sub, topic: topic}, nil
}
