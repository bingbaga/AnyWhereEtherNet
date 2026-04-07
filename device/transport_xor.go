package device

import (
	"sync/atomic"

	"github.com/KusakabeSi/EtherGuard-VPN/conn"
	"github.com/KusakabeSi/EtherGuard-VPN/mtypes"
	"github.com/KusakabeSi/EtherGuard-VPN/transport"
)

func (device *Device) LookupPeerByID(id mtypes.Vertex, endpoint conn.Endpoint) *Peer {
	device.peers.RLock()
	defer device.peers.RUnlock()

	if id != mtypes.NodeID_SuperNode {
		return device.peers.IDMap[id]
	}

	var fallback *Peer
	for _, peer := range device.peers.SuperPeer {
		if fallback == nil {
			fallback = peer
		}
		if endpoint == nil || peer.endpoint == nil {
			continue
		}
		if peer.endpoint.DstToString() == endpoint.DstToString() {
			return peer
		}
	}
	return fallback
}

func (peer *Peer) nextXORSendSeq() uint32 {
	return atomic.AddUint32(&peer.xorSendSeq, 1) - 1
}

func (peer *Peer) SendTransportBuffer(elem *QueueOutboundElement) error {
	packet, err := peer.device.transport.Encode(peer.device.ID, elem.Type, elem.TTL, peer.nextXORSendSeq(), elem.packet)
	if err != nil {
		return err
	}
	return peer.SendBuffer(packet)
}

type xorReplayFilter = transport.XORReplayFilter
