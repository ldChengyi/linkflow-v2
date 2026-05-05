package event

import "fmt"

const (
	EnvelopeSchemaFile = "events/envelope.schema.json"

	TelemetryReceivedV1SchemaFile = "events/device.telemetry.received.v1.schema.json"
)

func PayloadSchemaKey(eventType string, version int) string {
	return fmt.Sprintf("%s:v%d", eventType, version)
}
