package rabbitmq

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
)

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
