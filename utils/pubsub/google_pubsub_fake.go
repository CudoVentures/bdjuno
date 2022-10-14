package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type GoogleFakePubSubClient struct {
	*GooglePubSubClient
	topic     *pubsub.Topic
	AckCount  int
	NackCount int
}

func NewFakeGooglePubSubClient(ctx context.Context) (*GoogleFakePubSubClient, error) {
	server := pstest.NewServer()
	client, err := pubsub.NewClient(
		ctx,
		"p",
		option.WithEndpoint(server.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
	)
	if err != nil {
		return nil, err
	}

	topic, err := client.CreateTopic(ctx, "t")
	if err != nil {
		return nil, err
	}

	sub, err := client.CreateSubscription(ctx, "s", pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		return nil, err
	}
	sub.ReceiveSettings.Synchronous = true

	ps := &GoogleFakePubSubClient{
		&GooglePubSubClient{ctx: ctx, sub: sub},
		topic,
		0,
		0,
	}

	return ps, nil
}

func (c *GoogleFakePubSubClient) Publish(data []byte) {
	c.topic.Publish(c.ctx, &pubsub.Message{Data: data})
}

func (c *GoogleFakePubSubClient) Subscribe(callback func(*PubSubMsg)) error {
	return c.sub.Receive(c.ctx, func(_ context.Context, msg *pubsub.Message) {
		fakeMsg := &PubSubMsg{
			Data: msg.Data,
			Ack: func() {
				c.AckCount++
				msg.Ack()
			},
			Nack: func() {
				c.NackCount++
				msg.Nack()
			},
		}

		callback(fakeMsg)
	})
}
