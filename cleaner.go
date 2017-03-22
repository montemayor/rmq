package rmq

import "fmt"

// Cleaner is a utility class for doing housekeeping to remove abandoned records
// from RMQ within Redis. It is good practice to have at least one client
// periodically call Clean.
type Cleaner struct {
	connection *RedisConnection
}

// NewCleaner returns an initialized Cleaner object.
func NewCleaner(connection *RedisConnection) *Cleaner {
	return &Cleaner{connection: connection}
}

// Clean inspects the set of active connections and removes any connections
// it detects that are no longer alive. Further,it calls `CleanConnection` for
// any each connection that it purges.
func (cleaner *Cleaner) Clean() error {
	connectionNames := cleaner.connection.GetConnections()
	for _, connectionName := range connectionNames {
		connection := cleaner.connection.hijackConnection(connectionName)
		if connection.Check() {
			continue // skip active connections!
		}

		if err := cleaner.CleanConnection(nil); err != nil {
			return err
		}
	}

	return nil
}

// CleanConnection calls CleanQueue on any queues marked open by a passed in connection.
// If connection is nil, the connection held by the Cleaner will be cleaned.
func (cleaner *Cleaner) CleanConnection(connection *RedisConnection) error {
	if connection == nil {
		connection = cleaner.connection
	}
	queueNames := connection.GetConsumingQueues()
	for _, queueName := range queueNames {
		queue, ok := connection.OpenQueue(queueName).(*redisQueue)
		if !ok {
			return fmt.Errorf("rmq cleaner failed to open queue %s", queueName)
		}

		cleaner.CleanQueue(queue)
	}

	if !connection.Close() {
		return fmt.Errorf("rmq cleaner failed to close connection %s", connection)
	}

	if err := connection.CloseAllQueuesInConnection(); err != nil {
		return fmt.Errorf("rmq cleaner failed to close all queues %s %s", connection.String(), err)
	}

	// log.Printf("rmq cleaner cleaned connection %s", connection)
	return nil
}

// CleanQueue returns all unacknowledged messages in the provided queue back to
// the ready queue.
func (cleaner *Cleaner) CleanQueue(queue *redisQueue) {
	returned := queue.ReturnAllUnacked()
	queue.CloseInConnection()
	_ = returned
	// log.Printf("rmq cleaner cleaned queue %s %d", queue, returned)
}
