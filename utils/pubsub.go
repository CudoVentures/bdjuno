package utils

import (
	"context"

	"cloud.google.com/go/pubsub"
)

type PubSub interface {
	Subscribe(callback func(*Message)) error
}

type GooglePubSubClient struct {
	ctx   context.Context
	sub   *pubsub.Subscription
	topic *pubsub.Topic
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

func (c *GooglePubSubClient) Subscribe(callback func(*Message)) error {
	return c.sub.Receive(c.ctx, func(_ context.Context, msg *pubsub.Message) {
		callback(NewGooglePubSubMessage(msg))
	})
}

func (c *GooglePubSubClient) Publish(data []byte) *pubsub.PublishResult {
	return c.topic.Publish(c.ctx, &pubsub.Message{Data: data})
}

type Message struct {
	Data []byte
	Ack  func()
	Nack func()
}

func NewGooglePubSubMessage(msg *pubsub.Message) *Message {
	return &Message{
		Data: msg.Data,
		Ack:  msg.Ack,
		Nack: msg.Nack,
	}
}
