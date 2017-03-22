package rmq

// Consumer is the interface that must be implemented by users of RMQ for handling
// single messages (a delivery) at a time.
type Consumer interface {
	Consume(delivery Delivery)
}

// BatchConsumer is the interface that must be satisfied by users of RMQ if
// necessary or desired to handle batches of messages at a time.
type BatchConsumer interface {
	Consume(batch Deliveries)
}
