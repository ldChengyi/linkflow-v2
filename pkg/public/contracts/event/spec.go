package event

// Spec describes the shared metadata for one event contract.
type Spec struct {
	Type       string
	Version    int
	SchemaFile string
	Topic      string
}

func (s Spec) Key() string {
	return PayloadSchemaKey(s.Type, s.Version)
}

var DevicePropertyReported = Spec{
	Type:       TypeDevicePropertyReported,
	Version:    VersionDevicePropertyReported,
	SchemaFile: PropertyReportedV1SchemaFile,
	Topic:      TopicDeviceEventsV1,
}

var DevicePropertySetAcknowledged = Spec{
	Type:       TypeDevicePropertySetAcknowledged,
	Version:    VersionDevicePropertySetAcknowledged,
	SchemaFile: PropertySetAcknowledgedV1SchemaFile,
	Topic:      TopicDeviceEventsV1,
}

var DeviceConnected = Spec{
	Type:       TypeDeviceConnected,
	Version:    VersionDeviceConnected,
	SchemaFile: DeviceConnectedV1SchemaFile,
	Topic:      TopicDeviceEventsV1,
}

var DeviceDisconnected = Spec{
	Type:       TypeDeviceDisconnected,
	Version:    VersionDeviceDisconnected,
	SchemaFile: DeviceDisconnectedV1SchemaFile,
	Topic:      TopicDeviceEventsV1,
}

var DeviceConnectionChanged = Spec{
	Type:       TypeDeviceConnectionChanged,
	Version:    VersionDeviceConnectionChanged,
	SchemaFile: DeviceConnectionChangedV1SchemaFile,
	Topic:      TopicDeviceStateV1,
}

var DevicePropertyChanged = Spec{
	Type:       TypeDevicePropertyChanged,
	Version:    VersionDevicePropertyChanged,
	SchemaFile: DevicePropertyChangedV1SchemaFile,
	Topic:      TopicDeviceStateV1,
}

var specs = []Spec{
	DevicePropertyReported,
	DevicePropertySetAcknowledged,
	DeviceConnected,
	DeviceDisconnected,
	DeviceConnectionChanged,
	DevicePropertyChanged,
}

func Specs() []Spec {
	out := make([]Spec, len(specs))
	copy(out, specs)
	return out
}

func Lookup(eventType string, version int) (Spec, bool) {
	for _, s := range specs {
		if s.Type == eventType && s.Version == version {
			return s, true
		}
	}
	return Spec{}, false
}
