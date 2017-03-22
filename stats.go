package rmq

import (
	"fmt"
	"sort"
)

type ConnectionStat struct {
	Active       bool     `json:"active"`
	UnackedCount int      `json:"unacked"`
	Consumers    []string `json:"consumers"`
}

func (stat ConnectionStat) String() string {
	return fmt.Sprintf("[unacked:%d Consumers:%d]",
		stat.UnackedCount,
		len(stat.Consumers),
	)
}

type ConnectionStats map[string]ConnectionStat

type QueueStat struct {
	ReadyCount      int             `json:"ready"`
	RejectedCount   int             `json:"rejected"`
	ConnectionStats ConnectionStats `json:"connections"`
}

func NewQueueStat(readyCount, rejectedCount int) QueueStat {
	return QueueStat{
		ReadyCount:      readyCount,
		RejectedCount:   rejectedCount,
		ConnectionStats: ConnectionStats{},
	}
}

func (stat QueueStat) String() string {
	return fmt.Sprintf("[ready:%d rejected:%d conn:%s",
		stat.ReadyCount,
		stat.RejectedCount,
		stat.ConnectionStats,
	)
}

func (stat QueueStat) UnackedCount() int {
	unacked := 0
	for _, connectionStat := range stat.ConnectionStats {
		unacked += connectionStat.UnackedCount
	}
	return unacked
}

func (stat QueueStat) ConsumerCount() int {
	consumer := 0
	for _, connectionStat := range stat.ConnectionStats {
		consumer += len(connectionStat.Consumers)
	}
	return consumer
}

func (stat QueueStat) ConnectionCount() int {
	return len(stat.ConnectionStats)
}

type QueueStats map[string]QueueStat

type Stats struct {
	QueueStats       QueueStats      `json:"queues"`
	otherConnections map[string]bool // non consuming connections, Active or not
}

func NewStats() Stats {
	return Stats{
		QueueStats:       QueueStats{},
		otherConnections: map[string]bool{},
	}
}

func collectStats(queueList []string, mainConnection *RedisConnection) Stats {
	stats := NewStats()
	for _, queueName := range queueList {
		queue := mainConnection.openQueue(queueName)
		stats.QueueStats[queueName] = NewQueueStat(queue.ReadyCount(), queue.RejectedCount())
	}

	connectionNames := mainConnection.GetConnections()
	for _, connectionName := range connectionNames {
		connection := mainConnection.hijackConnection(connectionName)
		connectionActive := connection.Check()

		queueNames := connection.GetConsumingQueues()
		if len(queueNames) == 0 {
			stats.otherConnections[connectionName] = connectionActive
			continue
		}

		for _, queueName := range queueNames {
			queue := connection.openQueue(queueName)
			Consumers := queue.GetConsumers()
			openQueueStat, ok := stats.QueueStats[queueName]
			if !ok {
				continue
			}
			openQueueStat.ConnectionStats[connectionName] = ConnectionStat{
				Active:       connectionActive,
				UnackedCount: queue.UnackedCount(),
				Consumers:    Consumers,
			}
		}
	}

	return stats
}

func (stats ConnectionStats) sortedNames() []string {
	var keys []string
	for key := range stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (stats Stats) sortedQueueNames() []string {
	var keys []string
	for key := range stats.QueueStats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (stats Stats) sortedConnectionNames() []string {
	var keys []string
	for key := range stats.otherConnections {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ActiveSign(Active bool) string {
	if Active {
		return "✓"
	}
	return "✗"
}
