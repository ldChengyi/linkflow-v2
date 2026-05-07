package event

const (
	TypeDevicePropertyReported    = "device.property.reported"
	VersionDevicePropertyReported = 1
	TopicDeviceEventsV1           = "lf.v1.device.events"
)

type PropertyReportedPayload struct {
	DeviceID   string         `json:"device_id"`
	ProductKey string         `json:"product_key"`
	Protocol   string         `json:"protocol"`
	Properties map[string]any `json:"properties"`
	Raw        map[string]any `json:"raw,omitempty"`
}
