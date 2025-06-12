package main

import (
	"net"
	"fmt"
	"slices"
	"p2p/shared"
)

type Peer struct {
	port uint16
	peers []string 
	list net.Listener
	cons map[string]*net.Conn
}

func NewPeer(port uint16, peers []string) (*Peer, error) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &Peer{
		port: port,
		peers: peers,
		list: list,
	}, nil
}

func (p *Peer) Send(data []byte, to string) (int, error) {
	if !slices.Contains(p.peers, to) {
		return 0, fmt.Errorf("ERROR: peer not found\n")
	}
	conn, err := net.Dial("tcp", to)
	if err != nil {
		return 0, err
	}
	defer conn.Close()	
	nbw, err := conn.Write(data)
	if err != nil {
		return nbw, err
	}
	return nbw, nil
}

// func (p *Peer) Read() (*shared.Packet, error) {
// 	conn, err := p.list.Accept()
// 	if err != nil {
// 		return nil, err
// 	}
// 	addr := conn.RemoteAddr().String()
// 	if !slices.Contains
// }
