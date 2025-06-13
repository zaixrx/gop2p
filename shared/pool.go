package shared 

import (
	"net"
	"fmt"
	"sync"
	"slices"
)

type Pool struct {
	mux sync.Mutex
	id string
	host string
	peers []string 
}

func NewPool(id string, host *net.UDPAddr) *Pool {
	return &Pool{
		id: id,
		host: host.String(),
		peers: []string{host.String()},
	}
}

func (p *Pool) Add(peer string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	if exists, _ := p.PeerExists(peer); exists {
		return fmt.Errorf("ERROR: peer already exists in pool")
	}
	
	p.peers = append(p.peers, peer)
	return nil
}

func (p *Pool) Remove(peer string) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	if peer == p.host {
		return fmt.Errorf("ERROR: removing host from peer's pool requires removing the pool itself")
	}
	exists, index := p.PeerExists(peer)
	if !exists { 
		return fmt.Errorf("ERROR: peer not found")
	}
	n := len(p.peers)
	p.peers[index] = p.peers[n - 1]
	p.peers = p.peers[:n-1]
	return nil
}

func (p *Pool) PeerExists(peer string) (bool, int) {
	i := slices.Index(p.peers, peer) 
	return i != -1, i
}
