package pulsarmq

import (
	"context"
	"errors"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/cocktail828/go-tools/pkg/mq"
	"github.com/cocktail828/go-tools/z/variadic"
)

var (
	ErrMalformedMessage = errors.New("pulsar: malformed message")
)

type pulsarMQ struct {
	pulsar.Client
}

func NewClient(ctx context.Context, uri string) (mq.MQ, error) {
	cli, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:    uri,
		Logger: log.DefaultNopLogger(),
	})
	if err != nil {
		return nil, err
	}
	return pulsarMQ{cli}, nil
}

type consumerImpl struct {
	pulsar.Consumer
}

func (c consumerImpl) Receive(ctx context.Context) (mq.Message, error) {
	msg, err := c.Consumer.Receive(ctx)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c consumerImpl) Ack(msg mq.Message) error {
	if m, ok := msg.(pulsar.Message); ok {
		return c.Consumer.Ack(m)
	}
	return ErrMalformedMessage
}

func (c consumerImpl) AckCumulative(msg mq.Message) error {
	if m, ok := msg.(pulsar.Message); ok {
		return c.Consumer.AckCumulative(m)
	}
	return ErrMalformedMessage
}

func (c consumerImpl) Nack(msg mq.Message) {
	if m, ok := msg.(pulsar.Message); ok {
		c.Consumer.Nack(m)
	}
}

func (c consumerImpl) Close() error {
	c.Consumer.Close()
	return nil
}

func (mq pulsarMQ) Subscribe(subname, topic string, opts ...variadic.Option) (mq.Consumer, error) {
	v := inVariadic{variadic.Compose(opts...)}

	consumer, err := mq.Client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subname,
		Type:             v.SubscriptionType(),
		DLQ:              v.DLQPolicy(),
	})
	if err != nil {
		return nil, err
	}
	return consumerImpl{consumer}, nil
}

type producerImpl struct {
	pulsar.Producer
}

func (p producerImpl) Send(ctx context.Context, payload []byte) error {
	_, err := p.Producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: payload,
	})
	return err
}

func (p producerImpl) Close() error {
	p.Producer.Close()
	return nil
}

func (mq pulsarMQ) NewProducer(topic string, opts ...variadic.Option) (mq.Producer, error) {
	v := inVariadic{variadic.Compose(opts...)}

	p, err := mq.CreateProducer(pulsar.ProducerOptions{
		Topic:                   topic,
		DisableBatching:         v.DisableBatch(),
		DisableBlockIfQueueFull: true, // throw error instead of blocking
		CompressionType:         v.CompressType(),
		CompressionLevel:        v.CompressLevel(),
	})
	if err != nil {
		return nil, err
	}
	return producerImpl{p}, nil
}

func (mq pulsarMQ) Close() error {
	mq.Client.Close()
	return nil
}
