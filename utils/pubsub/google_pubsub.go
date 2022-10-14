package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
)

type GooglePubSubClient struct {
	ctx context.Context
	sub *pubsub.Subscription
}

func NewGooglePubSubClient(ctx context.Context, projectID string, subID string) (*GooglePubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	sub := client.Subscription(subID)
	sub.ReceiveSettings.MaxOutstandingMessages = 1
	sub.ReceiveSettings.NumGoroutines = 1

	return &GooglePubSubClient{ctx: ctx, sub: sub}, nil
}

func (c *GooglePubSubClient) Subscribe(callback func(*PubSubMsg)) error {
	return c.sub.Receive(c.ctx, func(_ context.Context, msg *pubsub.Message) {
		callback(NewGooglePubSubMessage(msg))
	})
}

func NewGooglePubSubMessage(msg *pubsub.Message) *PubSubMsg {
	return &PubSubMsg{
		Data: msg.Data,
		Ack:  msg.Ack,
		Nack: msg.Nack,
	}
}
