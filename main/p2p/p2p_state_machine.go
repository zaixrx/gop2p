// I still haven't decided how to handle to either listen for peers or handle sent messages
// I need to make jobState an official
// and don't forget the case for the sent messages needing to modify packet consume or not cosume

package P2P

import (
	"fmt"
	"sync"
	"time"
	"p2p/shared"
	machine "p2p/main/stateMachine"
)

type t_output struct {}

type t_state struct {
	nm *network_manager

	peersLock sync.Mutex
	peers map[string]t_conn // On New Connection

	sendQLock sync.Mutex
	sendQ []*P2PMessage // On New Send Message Request

	handlersLocks sync.Mutex
	handlers map[string]messageHandler // On New Received Messgage

	port uint16
	pool *shared.PublicPool
	// logger for erros
}

type messageHandler func(string, *shared.Packet)
var zeroMsgHandler messageHandler = func(_ string, _ *shared.Packet) { }

type stateMachine machine.StateMachine[t_state, t_output]
type jobState struct {
	state *t_state
	output *t_output
}

func CreateP2P(pool *shared.PublicPool, port uint16) *stateMachine {
	return (*stateMachine)(machine.NewStateMachine(t_state{
		pool: pool,
		port: port,
		peers: make(map[string]t_conn),
		handlers: map[string]messageHandler{
			"peerConnected": zeroMsgHandler,
			"peerDisconnected": zeroMsgHandler,
		},
		nm: NewNetworkManager(),
	}, ConnectToAvailablePeers))
}

func ConnectToAvailablePeers(s *t_state, o *t_output) (machine.StateJob[t_state, t_output], error) {	
	err := s.nm.Listen(fmt.Sprintf(":%d", s.port))
	if err != nil {
		return nil, err
	}

	js := &jobState{
		state: s,
		output: o,
	}
	for _, addr := range s.pool.PeerIPs {
		conn, err := s.nm.Connect(addr)
		if err != nil {
			continue
		}

		s.peersLock.Lock()
		s.peers[addr] = conn
		s.peersLock.Unlock()

		go js.HandlePeer(addr)
	}

	return HandleSentMessages, nil
}

func HandleSentMessages(s *t_state, o *t_output) (machine.StateJob[t_state, t_output], error) {
	limitter := time.Tick(time.Millisecond * time.Duration(1000 / 30))
	for {
		<-limitter
		for _, msg := range s.sendQ {
			to, _ := msg.packet.ReadString() // Handle errors? i don't fucking this so	
			conn := s.peers[to]
			conn.Write(msg)
		}
	}

}

func HandleAcceptConns(s *t_state, o *t_output) (machine.StateJob[t_state, t_output], error) {
	js := &jobState{
		state: s,
		output: o,
	}
	for {
		conn, err := s.nm.Accept()
		if err != nil {
			continue
		}

		s.peersLock.Lock()
		pid := conn.Address()
		s.peers[pid] = conn
		s.peersLock.Unlock()

		go js.HandlePeer(pid)
	}
}

func (js *jobState) HandlePeer(pid string) {
	conn := js.state.peers[pid]	
	
	for {
		msg, err := conn.Read()	
		if err != nil {
			js.DisconnectPeer(pid)
			return
		}

		from, msgType, packet, err := ExtractMessage(msg.packet)
		if err != nil {
			continue // TODO: error
		}

		handler, exists := js.state.handlers[msgType]
		if !exists {
			// fmt.Errorf("ERROR: invalid message type, handler doesn't exist")
			continue
		}

		handler(from, packet)
	}
}

func (js *jobState) DisconnectPeer(addr string) error {
	js.state.peersLock.Lock()
	defer js.state.peersLock.Unlock()

	_, exists := js.state.peers[addr]
	if !exists {
		return fmt.Errorf("ERROR: cannot delete unregister peer")
	}

	delete(js.state.peers, addr)

	return nil
}

func ExtractMessage(packet *P2PPacket) (string, string, *P2PPacket, error) {
	peer, err := packet.ReadString()
	if err != nil {
		return "", "", nil, err
	}
	msgType, err := packet.ReadString()
	if err != nil {
		return peer, "", nil, err
	}
	return peer, msgType, packet, nil
}

func (sm *stateMachine)On(msg string, handler messageHandler) {
	state := (*machine.StateMachine[t_state, t_output])(sm).GetState()
	
	state.peersLock.Lock()
	defer state.peersLock.Unlock()

	state.handlers[msg] = handler
}

func (sm *stateMachine)Send(to string, msg string, packet *shared.Packet) error {
	state := (*machine.StateMachine[t_state, t_output])(sm).GetState()

	if _, exists := state.peers[to]; !exists {
		return fmt.Errorf("ERROR: destination connection doesn't exist")
	}

	if _, exists := state.handlers[msg]; !exists {
		return fmt.Errorf("ERROR: can't send message with no handler")
	}

	prePacket := NewPacket()
	prePacket.WriteString(to)
	prePacket.WriteString(msg)
	prePacket.WriteBytesAfter(packet.GetBytes())
	
	state.sendQLock.Lock()
	state.sendQ = append(state.sendQ, &P2PMessage{
		packet: prePacket,
	})
	state.sendQLock.Unlock()	

	return nil
}
