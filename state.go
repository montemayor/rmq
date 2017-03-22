package rmq

//go:generate stringer -type=State

// State defines the status of an RMQ message.
type State int

const (
	// Unacked messages have been sent to a consumer, but the consumer has not
	// formally completed work on the message yet
	Unacked State = iota
	// Acked messages are messages that have been successfully processed by a
	// consumer
	Acked
	// Rejected messages are messages that have been marked by a consumer as
	// rejected - typically due to an error in processing.
	Rejected
	// Pushed messages are messages that have been sent to a different queue
	// by the consumer.
	Pushed
)
