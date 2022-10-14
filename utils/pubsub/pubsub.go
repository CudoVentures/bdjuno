package pubsub

type PubSubClient interface {
	Subscribe(callback func(*PubSubMsg)) error
}

type PubSubMsg struct {
	Data []byte
	Ack  func()
	Nack func()
}
