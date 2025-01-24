package mq

import (
	"context"

	"github.com/cocktail828/go-tools/z/variadic"
)

type Message interface {
	Topic() string
	Payload() []byte
}

type Consumer interface {
	Ack(Message) error
	AckCumulative(Message) error
	Nack(Message)
	Receive(context.Context) (Message, error)
	Close() error
}

type Producer interface {
	Topic() string
	Send(ctx context.Context, payload []byte) error
	Close() error
}

type MQ interface {
	Subscribe(subname, topic string, opts ...variadic.Option) (Consumer, error)
	NewProducer(topic string, opts ...variadic.Option) (Producer, error)
	Close() error
}
