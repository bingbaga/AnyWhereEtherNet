package device

func (device *Device) SetListenPort(port uint16) error {
	device.net.Lock()
	if device.net.port == port {
		device.net.Unlock()
		return nil
	}
	device.net.port = port
	device.net.Unlock()

	return device.BindUpdate()
}
