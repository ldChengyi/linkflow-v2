package event

const (
	TypeDeviceServiceCallAcknowledged    = "device.service.call.acknowledged"
	VersionDeviceServiceCallAcknowledged = 1
)

type ServiceCallAcknowledgedPayload struct {
	CommandID   string         `json:"command_id"`
	TenantID    string         `json:"tenant_id"`
	ProductID   string         `json:"product_id"`
	DeviceID    string         `json:"device_id"`
	TenantSlug  string         `json:"tenant_slug"`
	ProductKey  string         `json:"product_key"`
	DeviceSlug  string         `json:"device_slug"`
	Protocol    string         `json:"protocol"`
	ServiceName string         `json:"service_name"`
	Success     bool           `json:"success"`
	Code        string         `json:"code,omitempty"`
	Message     string         `json:"message,omitempty"`
	Output      map[string]any `json:"output,omitempty"`
	Raw         map[string]any `json:"raw,omitempty"`
}
