package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewClient(addr string) (*Client, error) {
	conn, err := amqp091.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("dialing rabbitmq: %w", err)
	}

	client := Client{conn: conn}

	if err := client.init(conn); err != nil {
		return nil, fmt.Errorf("failed to init: %w", err)
	}

	log.Println("RabbitMQ connnect successfully")
	return &client, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		log.Println("closing rabbitmq connection...")
		return err
	}

	return nil
}

func (c *Client) init(conn *amqp091.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	c.channel = ch
	return nil
}

func (c *Client) Publish(exchange, key string, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg := amqp091.Publishing{
		Timestamp:   time.Now(),
		ContentType: "application/json",
		Body:        data,
	}

	if err := c.channel.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}
