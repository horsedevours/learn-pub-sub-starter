package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	err := subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(b []byte) (T, error) {
		var data *T
		err := json.Unmarshal(b, &data)
		if err != nil {
			return *new(T), err
		}
		return *data, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	err := subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(b []byte) (T, error) {
		data := *new(T)
		decoder := gob.NewDecoder(bytes.NewReader(b))
		err := decoder.Decode(&data)
		if err != nil {
			return *new(T), err
		}
		return data, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	err = channel.Qos(10, 0, false)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			data, err := unmarshaller(delivery.Body)
			if err != nil {
				log.Printf("error unmarshalling delivery: %s\n", err)
			}
			ackType := handler(data)
			switch ackType {
			case Ack:
				err := delivery.Ack(false)
				if err != nil {
					log.Printf("Error acknowledging delivery: %v\n", err)
				}
			case NackDiscard:
				err := delivery.Nack(false, false)
				if err != nil {
					log.Printf("Error nacknowledging delivery: %v\n", err)
				}
			case NackRequeue:
				err := delivery.Nack(false, true)
				if err != nil {
					log.Printf("Error nacknowledging delivery: %v\n", err)
				}
			default:
				log.Printf("Uknown AckType: %d\n", ackType)
			}

		}
	}()

	return nil
}
