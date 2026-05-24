package event

const (
	TypeDevicePropertySetRequested       = "device.property.set.requested"
	VersionDevicePropertySetRequested    = 1
	TypeDevicePropertySetAcknowledged    = "device.property.set.acknowledged"
	VersionDevicePropertySetAcknowledged = 1
)

type PropertySetRequestedPayload struct {
	CommandID   string         `json:"command_id"`
	TenantID    string         `json:"tenant_id"`
	ProductID   string         `json:"product_id"`
	DeviceID    string         `json:"device_id"`
	TenantSlug  string         `json:"tenant_slug"`
	DeviceSlug  string         `json:"device_slug"`
	ProductKey  string         `json:"product_key"`
	Protocol    string         `json:"protocol"`
	Topic       string         `json:"topic"`
	Properties  map[string]any `json:"properties"`
	RequestedBy string         `json:"requested_by"`
}

type PropertySetAcknowledgedPayload struct {
	CommandID  string         `json:"command_id,omitempty"`
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
