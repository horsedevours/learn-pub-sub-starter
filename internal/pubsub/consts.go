package pubsub

type QueueType int

const (
	Transient QueueType = iota
	Durable
)

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)
