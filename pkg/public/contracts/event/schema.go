package event

import "fmt"

const (
	EnvelopeSchemaFile = "events/envelope.schema.json"

	PropertyReportedV1SchemaFile        = "events/device.property.reported.v1.schema.json"
	PropertySetRequestedV1SchemaFile    = "events/device.property.set.requested.v1.schema.json"
	PropertySetAcknowledgedV1SchemaFile = "events/device.property.set.acknowledged.v1.schema.json"
	DeviceConnectedV1SchemaFile         = "events/device.connection.connected.v1.schema.json"
	DeviceDisconnectedV1SchemaFile      = "events/device.connection.disconnected.v1.schema.json"
	DeviceConnectionChangedV1SchemaFile = "events/device.connection.changed.v1.schema.json"
	DevicePropertyChangedV1SchemaFile   = "events/device.property.changed.v1.schema.json"
	DeviceEventReceivedV1SchemaFile     = "events/device.event.received.v1.schema.json"
	DeviceEventReportedV1SchemaFile     = "events/device.event.reported.v1.schema.json"
	ServiceCallRequestedV1SchemaFile    = "events/device.service.call.requested.v1.schema.json"
	ServiceCallAcknowledgedV1SchemaFile = "events/device.service.call.acknowledged.v1.schema.json"
)

func PayloadSchemaKey(eventType string, version int) string {
	return fmt.Sprintf("%s:v%d", eventType, version)
}
