package adapters

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/kafka"
)

var eventMessagePublisher *MessagePublisher

const (
	EventProductCreated      = "ProductCreated"
	EventProductChanged      = "ProductChanged"
	EventProductPriceChanged = "ProductPriceChanged"
	EventProductUidChanged   = "ProductUidChanged"
	EventSkuAdded            = "SkuAdded"
	EventSkuChanged          = "SkuChanged"
	EventSkuUidChanged       = "SkuUidChanged"
)

type MessagePublisher struct {
	producer *kafka.Producer
}

type Payload interface {
	ToEvent(ctx context.Context) interface{}
}

func SetupMessagePublisher(kafkaConfig kafka.Config) error {
	if len(kafkaConfig.Brokers) == 0 {
		return nil
	}

	producer, err := kafka.NewProducer(kafkaConfig.Brokers, kafkaConfig.Topic, func(c *sarama.Config) {
		c.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
		c.Producer.Compression = sarama.CompressionGZIP     // Compress messages
		c.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	})

	if err != nil {
		return err
	}

	eventMessagePublisher = &MessagePublisher{
		producer: producer,
	}
	return nil
}

func (MessagePublisher) Close() {
	if eventMessagePublisher != nil {
		eventMessagePublisher.producer.Close()
	}
}

func (MessagePublisher) Publish(ctx context.Context, payload Payload, status string) error {
	if eventMessagePublisher == nil {
		return nil
	}
	m := map[string]interface{}{
		"authToken": behaviorlog.FromCtx(ctx).AuthToken,
		"requestId": behaviorlog.FromCtx(ctx).RequestID,
		"status":    status,
		"payload":   payload.ToEvent(ctx),
		"createdAt": time.Now().UTC(),
	}
	return eventMessagePublisher.producer.Send(m)
}
