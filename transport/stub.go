package transport

import (
	"fmt"

	"github.com/bingbaga/AnyWhereEtherNet/mtypes"
)

type stubFactory struct {
	name string
}

func (f stubFactory) Name() string {
	return f.name
}

func (f stubFactory) Validate(cfg mtypes.TransportConfig) error {
	return nil
}

func (f stubFactory) New(cfg mtypes.TransportConfig) (Protocol, error) {
	return nil, fmt.Errorf("transport protocol %s is not implemented yet", f.name)
}

func init() {
	Register(stubFactory{name: "tls_tunnel"})
	Register(stubFactory{name: "dtls_tunnel"})
}
