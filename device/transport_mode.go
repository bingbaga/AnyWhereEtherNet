package device

func (device *Device) usesNoiseTransport() bool {
	return false
}

func (device *Device) TransportProtocol() string {
	if device.transport == nil {
		return ""
	}
	return device.transport.Name()
}
