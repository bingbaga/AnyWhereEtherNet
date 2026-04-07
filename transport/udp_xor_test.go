package transport

import (
	"bytes"
	"testing"

	"github.com/KusakabeSi/EtherGuard-VPN/mtypes"
	"github.com/KusakabeSi/EtherGuard-VPN/path"
)

func TestXORPacketRoundTrip(t *testing.T) {
	proto, err := New(mtypes.TransportConfig{
		Protocol: "udp_xor",
		XOR: mtypes.TransportXORConfig{
			Key: "shared-secret",
		},
	})
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	originalPayload := []byte{1, 2, 3, 4, 5, 6}
	packet, err := proto.Encode(mtypes.Vertex(7), path.PingPacket, 9, 42, append([]byte(nil), originalPayload...))
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	decoded, err := proto.Decode(packet)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if decoded.SenderID != 7 {
		t.Fatalf("sender mismatch: got %v want %v", decoded.SenderID, 7)
	}
	if decoded.Usage != path.PingPacket {
		t.Fatalf("usage mismatch: got %v want %v", decoded.Usage, path.PingPacket)
	}
	if decoded.TTL != 9 {
		t.Fatalf("ttl mismatch: got %d want %d", decoded.TTL, 9)
	}
	if decoded.Sequence != 42 {
		t.Fatalf("seq mismatch: got %d want %d", decoded.Sequence, 42)
	}
	if !bytes.Equal(decoded.Payload, originalPayload) {
		t.Fatalf("payload mismatch: got %v want %v", decoded.Payload, originalPayload)
	}
}

func TestXORPacketRoundTripWithHeaderObfuscation(t *testing.T) {
	proto, err := New(mtypes.TransportConfig{
		Protocol: "udp_xor",
		XOR: mtypes.TransportXORConfig{
			Key:              "shared-secret",
			ObfuscateHeaders: true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	packet, err := proto.Encode(mtypes.Vertex(9), path.ServerUpdate, 3, 99, []byte("payload"))
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	if packet[6] == byte(path.ServerUpdate) {
		t.Fatalf("expected obfuscated header bytes")
	}

	decoded, err := proto.Decode(packet)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if decoded.SenderID != 9 || decoded.Usage != path.ServerUpdate || decoded.TTL != 3 || decoded.Sequence != 99 {
		t.Fatalf("decoded header mismatch: sender=%v usage=%v ttl=%d seq=%d", decoded.SenderID, decoded.Usage, decoded.TTL, decoded.Sequence)
	}
	if string(decoded.Payload) != "payload" {
		t.Fatalf("payload mismatch: got %q", string(decoded.Payload))
	}
}

func TestXORReplayFilter(t *testing.T) {
	var filter XORReplayFilter
	if !filter.ValidateCounter(10, 64) {
		t.Fatalf("expected first packet to pass")
	}
	if filter.ValidateCounter(10, 64) {
		t.Fatalf("expected duplicate packet to be rejected")
	}
	if !filter.ValidateCounter(11, 64) {
		t.Fatalf("expected newer packet to pass")
	}
	if !filter.ValidateCounter(9, 64) {
		t.Fatalf("expected in-window older packet to pass once")
	}
	if filter.ValidateCounter(9, 64) {
		t.Fatalf("expected duplicate older packet to be rejected")
	}
	if filter.ValidateCounter(0, 4) {
		t.Fatalf("expected packet outside replay window to be rejected")
	}
}
