package shared

import (
	"fmt"
	"net"
	"sync"
)

type ServerPool struct {
	mux sync.Mutex
	Id string
	HostID string
	PingChan chan struct{}
	Peers map[string]*net.UDPAddr
}

type PublicPool struct {
	Id string
	HostIP string
	PeerIPs []string
}

func NewPool(id string, host *net.UDPAddr) *ServerPool {
	pool := &ServerPool{
		Id: id,
		HostID: host.String(),
		Peers: make(map[string]*net.UDPAddr),
	}

	pool.Add(host)
	
	return pool
}

func (p *ServerPool) Add(peer *net.UDPAddr) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	peerID := peer.String()
	if p.PeerExists(peerID) {
		return fmt.Errorf("ERROR: peer already exists in pool")
	}
	p.Peers[peerID] = peer

	return nil
}

func (p *ServerPool) Remove(peerID string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

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

func (p *ServerPool) ToPublic() *PublicPool {
	keys := make([]string, len(p.Peers))
	i := 0
	for key := range p.Peers {
		keys[i] = key
		i++
	}
	host := p.Peers[p.HostID]
	return &PublicPool{
		Id: p.Id,
		HostIP: host.String(),
		PeerIPs: keys,
	}
}

func (p *ServerPool) Ping() {
	p.mux.Lock()
	defer p.mux.Unlock()
	fmt.Println("Ping!")
	p.PingChan<-struct{}{}
}
