package transport

import (
	"encoding/binary"
	"errors"
	"sync"

	"github.com/bingbaga/AnyWhereEtherNet/mtypes"
	"github.com/bingbaga/AnyWhereEtherNet/path"
)

const (
	UDPXORHeaderSize = 14
	xorMagic0        = byte('E')
	xorMagic1        = byte('X')
	xorVersion       = byte(1)

	xorFlagHeaderObfuscated = 1 << 0
)

var (
	ErrPacketTooSmall   = errors.New("udp_xor packet too small")
	ErrInvalidMagic     = errors.New("udp_xor invalid magic")
	ErrUnsupportedFlags = errors.New("udp_xor unsupported flags")
	ErrInvalidVersion   = errors.New("udp_xor invalid version")
)

type XORReplayFilter struct {
	sync.Mutex
	initialized bool
	highest     uint32
	bitmap      uint64
}

func (f *XORReplayFilter) ValidateCounter(counter uint32, window uint32) bool {
	if window == 0 {
		window = 64
	}
	if window > 64 {
		window = 64
	}

	f.Lock()
	defer f.Unlock()

	if !f.initialized {
		f.initialized = true
		f.highest = counter
		f.bitmap = 1
		return true
	}

	if counter > f.highest {
		diff := counter - f.highest
		if diff >= 64 {
			f.bitmap = 1
		} else {
			f.bitmap = (f.bitmap << diff) | 1
		}
		f.highest = counter
		return true
	}

	diff := f.highest - counter
	if diff >= window {
		return false
	}

	mask := uint64(1) << diff
	if f.bitmap&mask != 0 {
		return false
	}
	f.bitmap |= mask
	return true
}

func (f *XORReplayFilter) Reset() {
	f.Lock()
	defer f.Unlock()
	f.initialized = false
	f.highest = 0
	f.bitmap = 0
}

type udpXORFactory struct{}

type UDPXORProtocol struct {
	key              []byte
	obfuscateHeaders bool
}

func (udpXORFactory) Name() string {
	return "udp_xor"
}

func (udpXORFactory) Validate(cfg mtypes.TransportConfig) error {
	if cfg.XOR.Key == "" {
		return errors.New("Transport.XOR.Key is required when Transport.Protocol=udp_xor")
	}
	return nil
}

func (f udpXORFactory) New(cfg mtypes.TransportConfig) (Protocol, error) {
	if err := f.Validate(cfg); err != nil {
		return nil, err
	}
	return &UDPXORProtocol{
		key:              []byte(cfg.XOR.Key),
		obfuscateHeaders: cfg.XOR.ObfuscateHeaders,
	}, nil
}

func (p *UDPXORProtocol) Name() string {
	return "udp_xor"
}

func (p *UDPXORProtocol) Overhead() int {
	return UDPXORHeaderSize
}

func (p *UDPXORProtocol) xorBytes(b []byte, offset int) {
	if len(p.key) == 0 {
		return
	}
	for i := range b {
		b[i] ^= p.key[(offset+i)%len(p.key)]
	}
}

func (p *UDPXORProtocol) Encode(sender mtypes.Vertex, usage path.Usage, ttl uint8, seq uint32, payload []byte) ([]byte, error) {
	packet := make([]byte, UDPXORHeaderSize+len(payload))
	packet[0] = xorMagic0
	packet[1] = xorMagic1
	packet[2] = xorVersion

	flags := byte(0)
	binary.BigEndian.PutUint16(packet[4:6], uint16(sender))
	packet[6] = byte(usage)
	packet[7] = ttl
	binary.LittleEndian.PutUint32(packet[8:12], seq)

	if p.obfuscateHeaders {
		flags |= xorFlagHeaderObfuscated
		p.xorBytes(packet[4:UDPXORHeaderSize], 4)
	}
	packet[3] = flags

	copy(packet[UDPXORHeaderSize:], payload)
	p.xorBytes(packet[UDPXORHeaderSize:], UDPXORHeaderSize)
	return packet, nil
}

func (p *UDPXORProtocol) Decode(packet []byte) (Packet, error) {
	if len(packet) < UDPXORHeaderSize {
		return Packet{}, ErrPacketTooSmall
	}
	if packet[0] != xorMagic0 || packet[1] != xorMagic1 {
		return Packet{}, ErrInvalidMagic
	}
	if packet[2] != xorVersion {
		return Packet{}, ErrInvalidVersion
	}

	flags := packet[3]
	if flags&^xorFlagHeaderObfuscated != 0 {
		return Packet{}, ErrUnsupportedFlags
	}
	if flags&xorFlagHeaderObfuscated != 0 {
		p.xorBytes(packet[4:UDPXORHeaderSize], 4)
	}

	payload := packet[UDPXORHeaderSize:]
	p.xorBytes(payload, UDPXORHeaderSize)
	return Packet{
		SenderID: mtypes.Vertex(binary.BigEndian.Uint16(packet[4:6])),
		Usage:    path.Usage(packet[6]),
		TTL:      packet[7],
		Sequence: binary.LittleEndian.Uint32(packet[8:12]),
		Payload:  payload,
	}, nil
}

func init() {
	Register(udpXORFactory{})
}
