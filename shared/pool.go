package shared 

import (
	"fmt"
	"net"
)

type ServerPool struct {
	Id string
	HostID string
	PingChan chan struct{}
	Peers map[string]*net.UDPAddr
}

type PublicPool struct {
	Id string
	YourIP string
	HostIP string
	PeerIPs []string
}

func NewPool(id string, host *net.UDPAddr) *ServerPool {
	pool := &ServerPool{
		Id: id,
		HostID: host.String(),
		PingChan: make(chan struct{}),
		Peers: make(map[string]*net.UDPAddr),
	}

	pool.Add(host)
	
	return pool
}

func (p *ServerPool) Add(peer *net.UDPAddr) error {
	peerID := peer.String()
	if p.PeerExists(peerID) {
		return fmt.Errorf("ERROR: peer already exists in pool")
	}
	p.Peers[peerID] = peer

	return nil
}

func (p *ServerPool) Remove(peerID string) error {
	if peerID == p.HostID {
		return fmt.Errorf("ERROR: removing host from pool requires removing the pool itself")
	}

	if !p.PeerExists(peerID) { 
		return fmt.Errorf("ERROR: peer not found")
	}

	delete(p.Peers, peerID)
	
	return nil
}

func (p *ServerPool) PeerExists(peerID string) bool {
	_, exists := p.Peers[peerID]
	return exists
}

// WARINING: This function expects this existance of (you, host)
func (p *ServerPool) ToPublic(you *net.UDPAddr) *PublicPool {
	keys := make([]string, len(p.Peers))
	i := 0
	for key := range p.Peers {
		keys[i] = key
		i++
	}
	host := p.Peers[p.HostID]
	return &PublicPool{
		Id: p.Id,
		YourIP: you.String(),
		HostIP: host.String(),
		PeerIPs: keys,
	}
}
