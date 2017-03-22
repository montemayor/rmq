package rmq

import "testing"

func TestDeliveriesAck(t *testing.T) {
	d1 := NewTestDeliveryString("d1")
	d1.Ack()
	d2 := NewTestDeliveryString("d2")
	d2.Ack()
	d3 := NewTestDeliveryString("d3")

	deliveries := Deliveries{d1, d2, d3}

	result := deliveries.Ack()

	if result != 2 {
		t.Error("Unexpected successful ack count. Expected", 2, "; got", result)
	}
	if d1.State != Acked {
		t.Error("d1 should be acked. State =", d1.State)
	}
	if d2.State != Acked {
		t.Error("d2 should be acked. State =", d2.State)
	}
	if d3.State != Acked {
		t.Error("d3 should be acked. State =", d3.State)
	}
}

func TestDeliveriesReject(t *testing.T) {
	d1 := NewTestDeliveryString("d1")
	d1.Reject()
	d2 := NewTestDeliveryString("d2")
	d3 := NewTestDeliveryString("d3")

	deliveries := Deliveries{d1, d2, d3}
	result := deliveries.Reject()

	if result != 1 {
		t.Error("Unexpected successful ack count. Expected", 2, "; got", result)
	}
	if d1.State != Rejected {
		t.Error("d1 should be rejected. State =", d1.State)
	}
	if d2.State != Rejected {
		t.Error("d2 should be rejected. State =", d2.State)
	}
	if d3.State != Rejected {
		t.Error("d3 should be rejected. State =", d3.State)
	}
}
