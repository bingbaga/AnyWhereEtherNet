package main

import (
	"github.com/bingbaga/AnyWhereEtherNet/mtypes"
	"github.com/bingbaga/AnyWhereEtherNet/transport"
)

func validateTransportConfig(cfg mtypes.TransportConfig) error {
	return transport.Validate(cfg)
}
