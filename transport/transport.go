package transport

import (
	"fmt"
	"sort"

	"github.com/KusakabeSi/EtherGuard-VPN/mtypes"
	"github.com/KusakabeSi/EtherGuard-VPN/path"
)

type Packet struct {
	SenderID mtypes.Vertex
	Usage    path.Usage
	TTL      uint8
	Sequence uint32
	Payload  []byte
}

type Protocol interface {
	Name() string
	Overhead() int
	Encode(sender mtypes.Vertex, usage path.Usage, ttl uint8, seq uint32, payload []byte) ([]byte, error)
	Decode(packet []byte) (Packet, error)
}

type Factory interface {
	Name() string
	Validate(cfg mtypes.TransportConfig) error
	New(cfg mtypes.TransportConfig) (Protocol, error)
}

var registry = map[string]Factory{}

func Register(factory Factory) {
	registry[factory.Name()] = factory
}

func Validate(cfg mtypes.TransportConfig) error {
	factory, ok := registry[cfg.GetProtocol()]
	if !ok {
		return fmt.Errorf("unknown Transport.Protocol: %s", cfg.Protocol)
	}
	return factory.Validate(cfg)
}

func New(cfg mtypes.TransportConfig) (Protocol, error) {
	factory, ok := registry[cfg.GetProtocol()]
	if !ok {
		return nil, fmt.Errorf("unknown Transport.Protocol: %s", cfg.Protocol)
	}
	return factory.New(cfg)
}

func SupportedProtocols() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
