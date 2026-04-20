package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO: could try flat buffer serialize and protobuf
func (c *Client) PublishJSON(exchange, key string, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	return publish(c, exchange, key, data)
}

func (c *Client) PublishGob(exchange, key string, val any) error {
	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(val); err != nil {
		return fmt.Errorf("failed to encode gob: %w", err)
	}

	return publish(c, exchange, key, buf.Bytes())
}

func publish(c *Client, exchange, key string, bytes []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "application/gob",
		Body:        bytes,
	}

	if err := c.channel.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}
