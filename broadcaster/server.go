package main 

import (
	"fmt"
	"net"
	"p2p/shared"
	"strings"
	"sync"
)

type Peer struct {
	mux sync.Mutex
	pakt *shared.Packet
	addr *net.UDPAddr
}

func (p *Peer) String() string {
	return p.addr.String()
}

type Server struct {
	conn *net.UDPConn
	peers map[string]*Peer
}

func NewServer(conn *net.UDPConn) *Server {
	return &Server{
		conn: conn,
		peers: make(map[string]*Peer),
	}
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

func (s *Server) Read() (*Peer, error) {
	buff := make([]byte, 1024)
	n, addr, err := s.conn.ReadFromUDP(buff)
	if err != nil {
		return nil, err
	}
	p, ok := s.peers[addr.String()]
	if !ok {
		p = s.NewPeer(addr)
		s.peers[addr.String()] = p
	}
	p.pakt.Load(buff[:n])
	return p, nil
}

func (s *Server) NewPeer(addr *net.UDPAddr) *Peer {
	return &Peer{
		addr: addr,
		pakt: shared.NewPacket(addr.String()),
	}
}

func (s *Server) Send(dat []byte, to string) (int, error) {
	peer, found := s.peers[to];
	if !found {
		return 0, fmt.Errorf("ERROR: peer with id %s isn't found\n", to)
	}
	return s.conn.WriteTo(dat, peer.addr)
}
