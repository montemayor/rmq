package rmq

import (
	"bytes"
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

func CollectStats(queueList []string, mainConnection *RedisConnection) Stats {
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

func (stats Stats) String() string {
	var buffer bytes.Buffer

	for queueName, queueStat := range stats.QueueStats {
		buffer.WriteString(fmt.Sprintf("    queue:%s ready:%d rejected:%d unacked:%d Consumers:%d\n",
			queueName, queueStat.ReadyCount, queueStat.RejectedCount, queueStat.UnackedCount(), queueStat.ConsumerCount(),
		))

		for connectionName, connectionStat := range queueStat.ConnectionStats {
			buffer.WriteString(fmt.Sprintf("        connection:%s unacked:%d Consumers:%d Active:%t\n",
				connectionName, connectionStat.UnackedCount, len(connectionStat.Consumers), connectionStat.Active,
			))
		}
	}

	for connectionName, Active := range stats.otherConnections {
		buffer.WriteString(fmt.Sprintf("    connection:%s Active:%t\n",
			connectionName, Active,
		))
	}

	return buffer.String()
}

func (stats Stats) GetHtml(layout, refresh string) string {
	buffer := bytes.NewBufferString("<html>")

	if refresh != "" {
		buffer.WriteString(fmt.Sprintf(`<head><meta http-equiv="refresh" content="%s">`, refresh))
	}

	buffer.WriteString(`<body><table style="font-family:monospace">`)
	buffer.WriteString(`<tr><td>` +
		`queue</td><td></td><td>` +
		`ready</td><td></td><td>` +
		`rejected</td><td></td><td>` +
		`</td><td></td><td>` +
		`connections</td><td></td><td>` +
		`unacked</td><td></td><td>` +
		`Consumers</td><td></td></tr>`,
	)

	for _, queueName := range stats.sortedQueueNames() {
		queueStat := stats.QueueStats[queueName]
		connectionNames := queueStat.ConnectionStats.sortedNames()
		buffer.WriteString(fmt.Sprintf(`<tr><td>`+
			`%s</td><td></td><td>`+
			`%d</td><td></td><td>`+
			`%d</td><td></td><td>`+
			`%s</td><td></td><td>`+
			`%d</td><td></td><td>`+
			`%d</td><td></td><td>`+
			`%d</td><td></td></tr>`,
			queueName, queueStat.ReadyCount, queueStat.RejectedCount, "", len(connectionNames), queueStat.UnackedCount(), queueStat.ConsumerCount(),
		))

		if layout != "condensed" {
			for _, connectionName := range connectionNames {
				connectionStat := queueStat.ConnectionStats[connectionName]
				buffer.WriteString(fmt.Sprintf(`<tr style="color:lightgrey"><td>`+
					`%s</td><td></td><td>`+
					`%s</td><td></td><td>`+
					`%s</td><td></td><td>`+
					`%s</td><td></td><td>`+
					`%s</td><td></td><td>`+
					`%d</td><td></td><td>`+
					`%d</td><td></td></tr>`,
					"", "", "", ActiveSign(connectionStat.Active), connectionName, connectionStat.UnackedCount, len(connectionStat.Consumers),
				))
			}
		}
	}

	if layout != "condensed" {
		buffer.WriteString(`<tr><td>-----</td></tr>`)
		for _, connectionName := range stats.sortedConnectionNames() {
			Active := stats.otherConnections[connectionName]
			buffer.WriteString(fmt.Sprintf(`<tr style="color:lightgrey"><td>`+
				`%s</td><td></td><td>`+
				`%s</td><td></td><td>`+
				`%s</td><td></td><td>`+
				`%s</td><td></td><td>`+
				`%s</td><td></td><td>`+
				`%s</td><td></td><td>`+
				`%s</td><td></td></tr>`,
				"", "", "", ActiveSign(Active), connectionName, "", "",
			))
		}
	}

	buffer.WriteString(`</table></body></html>`)
	return buffer.String()
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
