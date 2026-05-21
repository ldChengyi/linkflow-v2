package event

const (
	TypeDeviceConnected    = "device.connection.connected"
	VersionDeviceConnected = 1

	TypeDeviceDisconnected    = "device.connection.disconnected"
	VersionDeviceDisconnected = 1
)

type ConnectedPayload struct {
	TenantID   string `json:"tenant_id"`
	ProductID  string `json:"product_id"`
	DeviceID   string `json:"device_id"`
	TenantSlug string `json:"tenant_slug"`
	ProductKey string `json:"product_key"`
	DeviceSlug string `json:"device_slug"`
	Protocol   string `json:"protocol"`
	Keepalive  int    `json:"keepalive,omitempty"`
}

type DisconnectedPayload struct {
	TenantID   string `json:"tenant_id"`
	ProductID  string `json:"product_id"`
	DeviceID   string `json:"device_id"`
	TenantSlug string `json:"tenant_slug"`
	ProductKey string `json:"product_key"`
	DeviceSlug string `json:"device_slug"`
	Protocol   string `json:"protocol"`
	Reason     string `json:"reason,omitempty"`
}
