package pubsub

import (
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
	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			var data *T
			json.Unmarshal(delivery.Body, &data)
			ackType := handler(*data)
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
