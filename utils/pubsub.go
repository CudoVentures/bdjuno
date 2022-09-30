package utils

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

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

type PubSub interface {
	Subscribe(callback func(*Message)) error
}

type GooglePubSub struct {
	ctx    context.Context
	client *pubsub.Client
	subID  string
}

func NewGooglePubSub(ctx context.Context, projectID string, subID string) (*GooglePubSub, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &GooglePubSub{
		ctx:    ctx,
		client: client,
		subID:  subID,
	}, nil
}

func (pb *GooglePubSub) Subscribe(callback func(*Message)) error {
	sub := pb.client.Subscription(pb.subID)
	sub.ReceiveSettings.MaxOutstandingMessages = 1

	return sub.Receive(pb.ctx, func(_ context.Context, msg *pubsub.Message) {
		fmt.Println(*msg.DeliveryAttempt)
		callback(NewGooglePubSubMessage(msg))
	})
}
