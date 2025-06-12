package main

import (
	"net"
	"fmt"
	"slices"
)

type Peer struct {
	port uint16
	peers []string 
	listener *net.Listener
}

func NewPeer(port uint16, peers []string) (*Peer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &Peer{
		port: port,
		peers: peers,
		listener: &listener,
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
