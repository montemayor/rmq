package rmq

// Deliveries represents a batch or slice of individual Delivery structs. This
// type includes additional convenience methods for managing a set of Delivery
// structs.
type Deliveries []Delivery

// Ack loops through the Delivery objects and Ack's (acknowledges) each
// Delivery. The function returns the number of failures encountered.
func (deliveries Deliveries) Ack() int {
	failedCount := 0
	for _, delivery := range deliveries {
		if !delivery.Ack() {
			failedCount++
		}
	}
	return failedCount
}

// Reject loops through the Delivery objects and Rejects each
// Delivery. The function returns the number of failures encountered.
func (deliveries Deliveries) Reject() int {
	failedCount := 0
	for _, delivery := range deliveries {
		if !delivery.Reject() {
			failedCount++
		}
	}
	return failedCount
}
