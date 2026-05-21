package event

const (
	TopicDeviceStateV1 = "lf.v1.device.state"

	TypeDeviceConnectionChanged    = "device.connection.changed"
	VersionDeviceConnectionChanged = 1

	TypeDevicePropertyChanged    = "device.property.changed"
	VersionDevicePropertyChanged = 1

	TypeDeviceEventReceived    = "device.event.received"
	VersionDeviceEventReceived = 1

	ConnectionStatusOnline  = "online"
	ConnectionStatusOffline = "offline"
)

// ConnectionChangedPayload is emitted by the device-event-processor after a
// device.connection.connected or device.connection.disconnected event has been applied to the
// authoritative state. It is the post-validation fan-out signal for downstream
// consumers (e.g. real-time UI, alerting).
type ConnectionChangedPayload struct {
	TenantID   string `json:"tenant_id"`
	ProductID  string `json:"product_id"`
	DeviceID   string `json:"device_id"`
	TenantSlug string `json:"tenant_slug"`
	ProductKey string `json:"product_key"`
	DeviceSlug string `json:"device_slug"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
}

// PropertyChangedPayload is emitted by the device-event-processor after a
// device.property.reported event has been validated and persisted. The
// Properties map contains only fields accepted by the current thing model.
type PropertyChangedPayload struct {
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	DeviceID   string         `json:"device_id"`
	TenantSlug string         `json:"tenant_slug"`
	ProductKey string         `json:"product_key"`
	DeviceSlug string         `json:"device_slug"`
	Properties map[string]any `json:"properties"`
}

// EventReceivedPayload is emitted by the device-event-processor after a
// device.event.reported event has been validated and persisted. Params contains
// only fields accepted by the current thing model event output definition.
type EventReceivedPayload struct {
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	DeviceID   string         `json:"device_id"`
	TenantSlug string         `json:"tenant_slug"`
	ProductKey string         `json:"product_key"`
	DeviceSlug string         `json:"device_slug"`
	EventName  string         `json:"event_name"`
	Params     map[string]any `json:"params"`
}
