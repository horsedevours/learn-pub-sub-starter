package pubsub

type QueueType int

const (
	Transient QueueType = iota
	Durable
)
