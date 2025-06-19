package P2P
/*
This is here for historical reasons only

package P2P

import (
	"fmt"
	"sync"
	"time"
	"context"
	"p2p/shared"
	machine "p2p/main/stateMachine"
)

type t_job_state struct {
	nm *network_manager

	peersLock sync.Mutex
	peers map[string]t_conn // On New Connection

	sendQLock sync.Mutex
	sendQ []*P2PMessage // On New Send Message Request

	handlersLocks sync.Mutex
	handlers map[string]messageHandler // On New Received Messgage

	port uint16
	pool *shared.PublicPool
	
	// TODO: logger for errors
	// and who the fuck cares about logging, if it works it works
	// if it deosn't? fuck you!
}

type messageHandler func(string, *shared.Packet)
var zeroMsgHandler messageHandler = func(_ string, _ *shared.Packet) { }

type stateMachine machine.StateMachine[t_job_state]

func CreateP2P(pool *shared.PublicPool, port uint16) *stateMachine {
	return (*stateMachine)(machine.NewStateMachine(t_job_state{
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

func ConnectToAvailablePeers(ctx context.Context, js *t_job_state) (machine.StateJob[t_job_state], error) {	
	err := js.nm.Listen(fmt.Sprintf(":%d", js.port))
	if err != nil {
		return nil, err
	}

	for _, addr := range js.pool.PeerIPs {
		conn, err := js.nm.Connect(addr)
		if err != nil {
			continue
		}

		js.peersLock.Lock()
		js.peers[addr] = conn
		js.peersLock.Unlock()

		go js.HandlePeer(addr)
	}

	return HandleAcceptConns, nil
}

func HandleAcceptConns(ctx context.Context, js *t_job_state) (machine.StateJob[t_job_state], error) {
	go HandleSentMessages(ctx, js)

	for {
		conn, err := js.nm.Accept()
		if err != nil {
			continue
		}

		js.peersLock.Lock()
		pid := conn.Address()
		js.peers[pid] = conn
		js.peersLock.Unlock()

		go js.HandlePeer(pid)
	}
}

func HandleSentMessages(ctx context.Context, js *t_job_state) {
	limitter := time.Tick(time.Millisecond * time.Duration(1000 / 30))
	for {
		<-limitter
		for _, msg := range js.sendQ {
			msg.packet.SetConsume(false)
			to, _ := msg.packet.ReadString() // Handle errors? i don't fucking this so
			conn := js.peers[to]
			conn.Write(msg)
		}
	}

}

func (js *t_job_state) HandlePeer(pid string) {
	conn := js.peers[pid]	
	
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

		handler, exists := js.handlers[msgType]
		if !exists {
			// fmt.Errorf("ERROR: invalid message type, handler doesn't exist")
			continue
		}

		handler(from, packet)
	}
}

func (js *t_job_state) DisconnectPeer(addr string) error {
	_, exists := js.peers[addr]
	if !exists {
		return fmt.Errorf("ERROR: cannot delete unregister peer")
	}

	js.peersLock.Lock()
	delete(js.peers, addr)
	js.peersLock.Unlock()

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

func (sm *stateMachine)Run(ctx context.Context) {
	(*machine.StateMachine[t_job_state])(sm).Run(ctx)
}

func (sm *stateMachine)On(msg string, handler func(from string, msg *P2PPacket)) {
	state := (*machine.StateMachine[t_job_state])(sm).GetState()
	state.peersLock.Lock()
	state.handlers[msg] = handler
	state.peersLock.Unlock()
}

func (sm *stateMachine)Send(to string, msg string, packet *shared.Packet) error {
	state := (*machine.StateMachine[t_job_state])(sm).GetState()

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
*/
