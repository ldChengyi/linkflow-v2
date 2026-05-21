package event

const (
	TypeDevicePropertySetAcknowledged    = "device.property.set.acknowledged"
	VersionDevicePropertySetAcknowledged = 1
)

type PropertySetAcknowledgedPayload struct {
	TenantID   string         `json:"tenant_id"`
	ProductID  string         `json:"product_id"`
	DeviceID   string         `json:"device_id"`
	TenantSlug string         `json:"tenant_slug"`
	DeviceSlug string         `json:"device_slug"`
	ProductKey string         `json:"product_key"`
	Protocol   string         `json:"protocol"`
	Success    bool           `json:"success"`
	Code       string         `json:"code,omitempty"`
	Message    string         `json:"message,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
}
