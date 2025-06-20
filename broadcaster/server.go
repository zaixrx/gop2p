package broadcaster

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/zaixrx/gop2p/logging"
	"github.com/zaixrx/gop2p/shared"
)

type messageHandler func(*shared.Packet, *net.UDPAddr, *Server) error 

type Server struct {
	conn *net.UDPConn

	poolsLock sync.Mutex
	pools map[string]*shared.ServerPool

	// handler map is static, so it's thread safe!
	handler map[shared.MessageType]messageHandler 
	
	PingPoolTimeout time.Duration
	Logger logging.Logger
}

func NewServer(pingPoolTimeout int) *Server {
	return &Server{
		PingPoolTimeout: time.Duration(pingPoolTimeout),
		Logger: logging.NewStdLogger(),
		pools: make(map[string]*shared.ServerPool),
		handler: map[shared.MessageType]messageHandler {
			shared.MessageRetrievePools: handlePoolRetreivalMessage,
			shared.MessageCreatePool: handlePoolCreateMessage,
			shared.MessageJoinPool: handlePoolJoinMessage,
			shared.MessagePoolPing: handlePoolPingMessage,
		},
	}
}

func (s *Server) Start(hostname string, port int) {
	addr := net.UDPAddr{
		IP: net.ParseIP(hostname),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		s.Logger.Error("Coulnd't listen on %d: %s\n", port, err) 
	}
	s.conn = conn
	s.Logger.Info("Server listening on port %d\n", port) 
}

func (s *Server) Stop() error {
	s.Logger.Info("Server stopped")
	return s.conn.Close()
}

func (s *Server) Write(b []byte, a *net.UDPAddr) (int, error) {
	//s.Logger.Debug("Wrote message of size %d to %s\n", len(b), a.String()) 
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

	//s.Logger.Debug("Received message %d from %s\n", msgTyp, addr.String())

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

func (s *Server) monitorPool(poolID string) {
	pool, err := s.GetPool(poolID)
	if err != nil {
		return 
	}

	s.Logger.Debug("Monitoring Pool %v\n", pool.PingChan) 
	for {
		select {
		case <-time.After(s.PingPoolTimeout * time.Second):
			s.Logger.Debug("Pool %s timedout\n", poolID)

			packet := shared.NewPacket()
			packet.WriteByte(byte(shared.MessagePoolPingTimeout))

			s.Write(packet.GetBytes(), pool.Peers[pool.HostID])

			s.poolsLock.Lock()
			delete(s.pools, poolID)
			s.poolsLock.Unlock()

			return
		case <-pool.PingChan:
			continue
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
