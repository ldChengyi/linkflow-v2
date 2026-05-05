package event

const (
	TypeDeviceTelemetryReceived    = "device.telemetry.received"
	VersionDeviceTelemetryReceived = 1
	TopicDeviceEventsV1            = "lf.v1.device.events"
)

type TelemetryReceivedPayload struct {
	DeviceID   string         `json:"device_id"`
	ProductKey string         `json:"product_key"`
	Protocol   string         `json:"protocol"`
	Metrics    map[string]any `json:"metrics"`
	Raw        map[string]any `json:"raw,omitempty"`
}
