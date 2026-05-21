package event

const (
	TypeDeviceEventReported    = "device.event.reported"
	VersionDeviceEventReported = 1
)

type EventReportedPayload struct {
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	DeviceID   string         `json:"device_id"`
	TenantSlug string         `json:"tenant_slug"`
	ProductKey string         `json:"product_key"`
	DeviceSlug string         `json:"device_slug"`
	Protocol   string         `json:"protocol"`
	EventName  string         `json:"event_name"`
	Params     map[string]any `json:"params"`
	Raw        map[string]any `json:"raw,omitempty"`
}
