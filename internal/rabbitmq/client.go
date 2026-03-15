package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

func (c *Client) PublishJSON(exchange, key string, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "application/json",
		Body:        data,
	}

	if err := c.channel.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}

func (c *Client) PublishGob(exchange, key string, val any) error {
	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(val); err != nil {
		return fmt.Errorf("failed to encode gob: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg := amqp.Publishing{
		Timestamp:   time.Now(),
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	}

	if err := c.channel.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

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

func SubscribeJSON[T any](c *Client, exchange, queue, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	return subscribe(c, exchange, queue, key, queueType, handler, func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		if err != nil {
			return target, err
		}

		return target, nil
	})
}

func SubscribeGob[T any](c *Client, exchange, queue, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	return subscribe(c, exchange, queue, key, queueType, handler, func(b []byte) (T, error) {
		var target T

		decoder := gob.NewDecoder(bytes.NewBuffer(b))
		if err := decoder.Decode(&target); err != nil {
			return target, err
		}

		return target, nil
	})
}

func subscribe[T any](c *Client, exchange, queue, key string, queueType SimpleQueueType, handler func(T) AckType, unmarshaller func([]byte) (T, error)) error {
	_, err := c.DeclareAndBind(exchange, queue, key, queueType)
	if err != nil {
		return err
	}

	messages, err := c.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume queue %v: %w", queue, err)
	}

	go func() {
		for message := range messages {
			data, err := unmarshaller(message.Body)
			if err != nil {
				log.Printf("unmarshal error: %v+\n", err)
				if err := message.Nack(false, false); err != nil {
					log.Printf("nack error: %+v", err)
				}
				continue
			}

			switch handler(data) {
			case Ack:
				if err := message.Ack(false); err != nil {
					log.Printf("ack error: %+v", err)
				}

			case NackRequeue:
				if err := message.Nack(false, true); err != nil {
					log.Printf("nack error: %+v", err)
				}

			case NackDiscard:
				if err := message.Nack(false, false); err != nil {
					log.Printf("nack error: %+v", err)
				}
			}
		}
	}()

	return nil
}
