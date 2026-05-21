package event

const (
	TypeDevicePropertyReported    = "device.property.reported"
	VersionDevicePropertyReported = 1
	TopicDeviceEventsV1           = "lf.v1.device.events"
)

type PropertyReportedPayload struct {
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	DeviceID   string         `json:"device_id"`
	TenantSlug string         `json:"tenant_slug"`
	DeviceSlug string         `json:"device_slug"`
	ProductKey string         `json:"product_key"`
	Protocol   string         `json:"protocol"`
	Properties map[string]any `json:"properties"`
	Raw        map[string]any `json:"raw,omitempty"`
}
