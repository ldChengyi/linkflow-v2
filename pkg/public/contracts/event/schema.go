package event

import "fmt"

const (
	EnvelopeSchemaFile = "events/envelope.schema.json"

	PropertyReportedV1SchemaFile        = "events/device.property.reported.v1.schema.json"
	PropertySetAcknowledgedV1SchemaFile = "events/device.property.set.acknowledged.v1.schema.json"
)

func PayloadSchemaKey(eventType string, version int) string {
	return fmt.Sprintf("%s:v%d", eventType, version)
}
