package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	var queue amqp.Queue
	if simpleQueueType == 1 {
		queue, err = channel.QueueDeclare(queueName, true, false, false, false, args)
	} else {
		queue, err = channel.QueueDeclare(queueName, false, true, true, false, args)
	}
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	channel.QueueBind(queueName, key, exchange, false, nil)

	return channel, queue, nil
}
