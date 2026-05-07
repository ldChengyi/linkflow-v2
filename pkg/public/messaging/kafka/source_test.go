package kafka

import (
	"testing"

	kafkago "github.com/segmentio/kafka-go"
)

func TestNewSourceValidatesRequiredOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  Options
	}{
		{name: "brokers", opt: Options{Topic: "events", GroupID: "consumer"}},
		{name: "topic", opt: Options{Brokers: []string{"127.0.0.1:19092"}, GroupID: "consumer"}},
		{name: "group", opt: Options{Brokers: []string{"127.0.0.1:19092"}, Topic: "events"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewSource(tt.opt); err == nil {
				t.Fatal("NewSource() error = nil, want validation error")
			}
		})
	}
}

func TestDeliveryMessageCopiesKafkaFields(t *testing.T) {
	delivery := Delivery{
		msg: kafkago.Message{
			Key:   []byte("k1"),
			Value: []byte("v1"),
			Headers: []kafkago.Header{
				{Key: "event_type", Value: []byte("device.property.reported")},
			},
		},
	}

	got := delivery.Message()
	if string(got.Key) != "k1" {
		t.Fatalf("key = %q", got.Key)
	}
	if string(got.Value) != "v1" {
		t.Fatalf("value = %q", got.Value)
	}
	if got.Headers["event_type"] != "device.property.reported" {
		t.Fatalf("event_type header = %q", got.Headers["event_type"])
	}

	got.Key[0] = 'x'
	got.Value[0] = 'x'
	if string(delivery.msg.Key) != "k1" {
		t.Fatalf("delivery key was mutated to %q", delivery.msg.Key)
	}
	if string(delivery.msg.Value) != "v1" {
		t.Fatalf("delivery value was mutated to %q", delivery.msg.Value)
	}
}
