package main

import (
	"github.com/KusakabeSi/EtherGuard-VPN/mtypes"
	"github.com/KusakabeSi/EtherGuard-VPN/transport"
)

func validateTransportConfig(cfg mtypes.TransportConfig) error {
	return transport.Validate(cfg)
}
