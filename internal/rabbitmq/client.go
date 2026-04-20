package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Transient = iota
	Durable
)

type AckType int

const (
	Ack = iota
	NackRequeue
	NackDiscard
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewClient(addr string) (*Client, error) {
	conn, err := amqp.Dial(addr)
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

func (c *Client) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	c.channel = ch
	return nil
}

func (c *Client) DeclareAndBind(exchange, queueName, key string, queueType SimpleQueueType) (amqp.Queue, error) {
	q, err := c.channel.QueueDeclare(queueName, queueType == Durable, queueType == Transient, queueType == Transient, false, amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := c.channel.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	return q, nil
}
