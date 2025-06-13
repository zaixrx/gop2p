package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"p2p/shared"
)

type MessageHandler func(*shared.Packet, *net.UDPAddr, *Server) error 

type Server struct {
	mux sync.Mutex
	conn *net.UDPConn
	pools map[string]*Pool
	handler map[shared.MessageType]MessageHandler 
}

func NewServer(conn *net.UDPConn) *Server {
	return &Server{
		conn: conn,
		pools: make(map[string]*Pool), 
		handler: map[shared.MessageType]MessageHandler {
			shared.MessageJoinPool: HandlePoolJoinMessage,
			shared.MessageLeavePool: HandlePoolLeaveMessage,
			shared.MessageCreatePool: HandlePoolCreateMessage,
			shared.MessageDeletePool: HandlePoolDeleteMessage,
			shared.MessageRetreivePools: HandlePoolRetreivalMessage,
		},
	}
}

func (s *Server) Listen() error {
	buff := make([]byte, 1024)
	n, addr, err := s.conn.ReadFromUDP(buff)
	if err != nil {
		return err
	}

	if n == 0 {
		log.Printf("Recevied empty buffer\n")
		return nil
	}

	err = s.handleMessage(buff[:n], addr)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Write(b []byte, a *net.UDPAddr) (int, error) {
	return s.conn.WriteToUDP(b, a)
}

func (s *Server) handleMessage(dat []byte, addr *net.UDPAddr) error {
	packet := shared.NewPacket()
	packet.Load(dat)

	msgTyp, err := packet.ReadByte()
	if err != nil {
		return err
	}

	log.Printf("Received message %d from %s\n", msgTyp, addr.String())

	handler, ok := s.handler[shared.MessageType(msgTyp)]
	if !ok {
		return fmt.Errorf("ERROR: unknown message type")
	}

	return handler(packet, addr, s)
}

func (s *Server) GetPool(id string) (*Pool, error) {
	pool, ok := s.pools[id]
	if !ok {
		return nil, fmt.Errorf("ERROR: pool with id %s doesn't exist", id)
	}
	return pool, nil 
}
