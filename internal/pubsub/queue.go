package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	var queue amqp.Queue
	if simpleQueueType == 1 {
		queue, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	} else {
		queue, err = channel.QueueDeclare(queueName, false, true, true, false, nil)
	}
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	channel.QueueBind(queueName, key, exchange, false, nil)

	return channel, queue, nil
}
