package main
// 
// import (
// 	"net"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"
// 	"slices"
// 	"context"
// 	"p2p/shared"
// )
// 
// //type StateAction[State any] func(context.Context, *State) (StateAction[State], error)
// 
// type P2PState struct {
// 	mux sync.Mutex
// 	Pool *shared.PublicPool
// 
// 	port uint16
// 	listener *net.Listener
// 	conns map[string]*net.Conn
// 	messagesToSend []*P2PMessage
// 	messages map[string]func(*shared.Packet)
// }
// 
// type P2PMessage struct {
// 	to string
// 	messageID string
// 	data *shared.Packet 
// }
// 
// func CreateP2PStage(pool *shared.PublicPool, port uint16) *StateContext[P2PState] {
// 	sc := NewStateContext(ConnectToPeers)
// 
// 	sc.state.Pool = pool
// 	sc.state.port = port
// 	sc.state.conns = make(map[string]*net.Conn)
// 	sc.state.messagesToSend = make([]*P2PMessage, 0)
// 	sc.state.messages = make(map[string]func(*shared.Packet))
// 
// 	return sc
// }
// 
// func (s *P2PState) On(messageID string, action func(*shared.Packet)) {
// 	s.mux.Lock()
// 	defer s.mux.Unlock()
// 
// 	s.messages[messageID] = action
// }
// 
// func ConnectToPeers(ctx context.Context, s *P2PState) (StateAction[P2PState], error) {
// 	s.mux.Lock()
// 	defer s.mux.Unlock()
// 
// 	log.Println("Connecting to other peers")
// 
// 	list, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
// 	if err != nil {
// 		return nil, err
// 	}
// 	s.listener = &list
// 
// 	for _, peerIP := range s.Pool.PeerIPs {
// 		s.addPeer(peerIP)
// 	}
// 
// 	if len(s.conns) < len(s.Pool.PeerIPs) - 1 {
// 		return nil, fmt.Errorf("Failed to connect to any peer\n", )
// 	}
// 
// 	go handleSendOps(ctx, s)
// 	return handleListen, nil
// }
// 
// func (s *P2PState) Send(to string, messageID string, data *shared.Packet) error {
// 	s.mux.Lock()
// 	defer s.mux.Unlock()
// 
// 	_, exists := s.conns[to]
// 	if !exists {
// 		return fmt.Errorf("ERROR: target peer doesn't exist in the current pool")
// 	}
// 	_, exists = s.messages[messageID]
// 	if !exists {
// 		return fmt.Errorf("ERROR: attempting to send a message that doesn't have a handler")
// 	}
// 	s.messagesToSend = append(s.messagesToSend, &P2PMessage{to, messageID, data})
// 	return nil
// }
// 
// func (s *P2PState) addPeer(peerIP string) error {
// 	s.mux.Lock()
// 	defer s.mux.Unlock()
// 
// 	var (
// 		conn net.Conn
// 		err error
// 	)
// 
// 	log.Println("YO new peer")
// 
// 	if !slices.Contains(s.Pool.PeerIPs, peerIP) {
// 		s.Pool.PeerIPs = append(s.Pool.PeerIPs, peerIP)
// 	}
// 
// 	for i := range 10 {
// 		if peerIP == s.Pool.YourIP {
// 			continue
// 		}
// 
// 		for i < shared.MaxPCR {
// 			conn, err = net.Dial("tcp", peerIP)
// 			if err != nil {
// 				continue
// 			}
// 			s.conns[peerIP] = &conn
// 			return nil
// 		}
// 	}
// 
// 	return fmt.Errorf("ERROR: failed to connect to %s\n", peerIP)
// }
// 
// func handleSendOps(ctx context.Context, state *P2PState) (StateAction[P2PState], error) {
// 	limitter := time.Tick(time.Millisecond * 1000 / 30)	
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil, nil
// 		case <-limitter:
// 			for _, msg := range state.messagesToSend {
// 				conn, exists := state.conns[msg.to]
// 				if !exists {
// 					// TODO: stream custom error
// 					continue
// 				}
// 				(*conn).Write(msg.data.GetBytes())
// 			}
// 		}	
// 	}
// }
// 
// func handleListen(ctx context.Context, s *P2PState) (StateAction[P2PState], error) {
// 	buff := make([]byte, 1024)
// 	packet := shared.NewPacket()
// 
// 	for {
// 		select {
// 			case <-ctx.Done():
// 				return nil, nil
// 			default:
// 				conn, err := (*s.listener).Accept()
// 				if err != nil {
// 					// TODO: stream the error
// 					continue
// 				}
// 				nbr, err := conn.Read(buff)
// 				if err != nil {
// 					// TODO: stream the error
// 					continue
// 				}
// 				peerIP := conn.RemoteAddr().String()
// 				if !slices.Contains(s.Pool.PeerIPs, peerIP) {
// 					s.addPeer(peerIP)
// 				}
// 				packet.Load(buff[:nbr])
// 				defer packet.Flush()
// 				msgHeader, err := packet.ReadString()
// 				if err != nil {
// 					// TODO: stream the error
// 					continue
// 				}
// 				handler, exists := s.messages[msgHeader]
// 				if !exists {
// 					// TODO: stream the custom error
// 					continue
// 				}
// 				handler(packet)
// 		}
// 	}
// }
