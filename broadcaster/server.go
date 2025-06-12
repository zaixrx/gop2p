package main 

import (
	"fmt"
	"net"
	"p2p/shared"
	"strings"
	"sync"
	"time"
	"log"
)

type Peer struct {
	mux sync.Mutex
	ping chan struct{}
	pakt *shared.Packet
	addr *net.UDPAddr
}

func (p *Peer) String() string {
	return p.addr.String()
}

func NewPeer(addr *net.UDPAddr) *Peer {
	return &Peer{
		addr: addr,
		ping: make(chan struct{}),
		pakt: shared.NewPacket(addr.String()),
	}
}

type Server struct {
	mux sync.Mutex
	conn *net.UDPConn
	peers map[string]*Peer
}

func NewServer(conn *net.UDPConn) *Server {
	return &Server{
		conn: conn,
		peers: make(map[string]*Peer),
	}
}

func (s *Server) Read() (*Peer, error) {
	buff := make([]byte, 1024)
	n, addr, err := s.conn.ReadFromUDP(buff)
	if err != nil {
		return nil, err
	}
	p, ok := s.peers[addr.String()]
	if !ok {
		p = s.addPeer(addr)
	}
	p.pakt.Load(buff[:n])
	return p, nil
}

func (s *Server) Send(dat []byte, to string) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	peer, found := s.peers[to];
	if !found {
		return 0, fmt.Errorf("ERROR: peer with id %s isn't found\n", to)
	}
	return s.conn.WriteTo(dat, peer.addr)
}

func (s *Server) addPeer(addr *net.UDPAddr) *Peer {
	s.mux.Lock()
	defer s.mux.Unlock()

	peer := NewPeer(addr)
	s.peers[peer.String()] = peer 

	go func() {
		for { 
			select {
			case <-time.After(shared.MaxPingTimeout * time.Second):
				s.removePeer(peer.String())
				return
			case <-peer.ping:
			}
		}
	}()
	
	return peer
}

func (s *Server) removePeer(id string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.peers[id]; !ok {
		return fmt.Errorf("ERROR: cannot delete non existing peer\n")
	}
	delete(s.peers, id)
	log.Printf("%s disconnected\n", id)
	return nil
}

func (s *Server) PeersString() string {
	var (
		keys = make([]string, len(s.peers))
		i = 0
	)
	for key := range s.peers {
		keys[i] = key
		i++
	}
	return strings.Join(keys, " ")
}
