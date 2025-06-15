package main

import (
	"fmt"
	"log"
	"net"
	"p2p/shared"
	"sync"
	"time"
)

type MessageHandler func(*shared.Packet, *net.UDPAddr, *Server) error 

type Server struct {
	conn *net.UDPConn

	mux sync.Mutex
	pools map[string]*shared.ServerPool

	handler map[shared.MessageType]MessageHandler 
}

func NewServer(conn *net.UDPConn) *Server {
	return &Server{
		conn: conn,
		pools: make(map[string]*shared.ServerPool),
		handler: map[shared.MessageType]MessageHandler {
			shared.MessageRetrievePools: HandlePoolRetreivalMessage,
			shared.MessageCreatePool: HandlePoolCreateMessage,
			shared.MessageJoinPool: HandlePoolJoinMessage,
			shared.MessagePoolPing: HandlePoolPingMessage,
		},
	}
}

func (s *Server) Write(b []byte, a *net.UDPAddr) (int, error) {
	return s.conn.WriteToUDP(b, a)
}

func (s *Server) Listen() error {
	buff := make([]byte, 1024)
	n, addr, err := s.conn.ReadFromUDP(buff)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	go s.handleMessage(buff[:n], addr)
	
	return nil
}

func (s *Server) handleMessage(dat []byte, addr *net.UDPAddr) error {
	packet := shared.NewPacket()
	packet.Load(dat)

	msgTyp, err := packet.ReadByte()
	if err != nil {
		s.reportError(err, addr)
		return nil
	}

	log.Printf("Received message %d from %s\n", msgTyp, addr.String())

	handler, ok := s.handler[shared.MessageType(msgTyp)]
	if !ok {
		s.reportError(fmt.Errorf("ERROR: unknown message type"), addr)
		return nil
	}

	err = handler(packet, addr, s)
	if err != nil {
		serr := s.reportError(err, addr)
		return serr 
	}

	return nil
}

func (s *Server) reportError(err error, to *net.UDPAddr) error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageError))
	packet.WriteString(err.Error())
	_, serr := s.Write(packet.GetBytes(), to)
	return serr 
}

func (s *Server) MonitorPool(poolID string) {
	pool, err := s.GetPool(poolID)
	if err != nil {
		return 
	}

	log.Printf("Monitoring Pool %v\n", pool.PingChan) 

	for {
		select {
		case <-pool.PingChan:
			continue
		case <-time.After(shared.PoolPingTimeout * time.Second):
			s.mux.Lock()
			defer s.mux.Unlock()
			
			log.Printf("Pool %s host's timedout\n", poolID)

			packet := shared.NewPacket()
			packet.WriteByte(byte(shared.MessagePoolPingTimeout))

			s.Write(packet.GetBytes(), pool.Peers[pool.HostID])

			delete(s.pools, poolID)
			return
		}
	}
}

func (s *Server) GetPool(poolID string) (*shared.ServerPool, error) {
	pool, ok := s.pools[poolID]
	if !ok {
		return nil, fmt.Errorf("ERROR: pool with id %s doesn't exist", poolID)
	}
	return pool, nil 
}
