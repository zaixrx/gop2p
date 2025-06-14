package main

import (
	"net"
	"log"
	"time"
	"slices"
	"context"
	"strconv"
	"strings"
	"p2p/shared"
)

//type StateAction[State any] func(state *State) (StateAction[State], error)

type State struct {
	br *Broadcast
	conn net.Conn
	packet *shared.Packet
	messageHandlers map[shared.MessageType]StateAction[State]

	pools []string
	currentPool *shared.PublicPool
}

func CreateBroadcastingStage(backgroundTask StateAction[State]) *StateContext[State] {
	return NewStateContext(ConnectToBroadcaster, backgroundTask)
}

func ConnectToBroadcaster(ctx context.Context, state *State) (StateAction[State], error) {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}

	state.conn = conn
	state.br = NewBroadcast(conn)
	state.packet = shared.NewPacket()
	state.pools = make([]string, 0)
	state.messageHandlers = map[shared.MessageType]StateAction[State]{
		shared.MessageRetrievePools: HandleRetreivePool,
		shared.MessageJoinPool: HandlePoolJoin,
		shared.MessagePoolPingTimeout: HandlePoolPingTimeout,	
	}

	return SendRetrievePools, nil 
}

func SendRetrievePools(ctx context.Context, state *State) (StateAction[State], error) {
	log.Println("Write Retreiving Pools")
	err := state.br.SendPoolRetreivalMessage()
	if err != nil {
		return nil, err
	}
	return state.Listen([]shared.MessageType{shared.MessageRetrievePools})
}

func HandleRetreivePool(ctx context.Context, state *State) (StateAction[State], error) {
	log.Println("Read Retreiving Pools")
	str, _ := state.packet.ReadString()
	state.pools = strings.Split(str, " ") 

	log.Println(state.pools)
	return state.Listen([]shared.MessageType{shared.MessageJoinPool})
}

func HandlePoolJoin(ctx context.Context, state *State) (StateAction[State], error) {
	log.Println("Read Joining Pool")
	pool, err := state.packet.ReadPool()
	if err != nil {
		return nil, err
	}
	state.currentPool = pool
	go func(ctx context.Context) {
		limitter := time.Tick(1000 / shared.PoolPingTicks * time.Millisecond)
		for {
			<-limitter
			if state.currentPool == nil {
				break
			}
			select {
			case <-ctx.Done():
				return
			default:
				state.br.SendPoolPingMessage(state.currentPool.Id)
			}
		}
	}(ctx)
	return state.Listen([]shared.MessageType{shared.MessagePoolPingTimeout})
}

func HandlePoolPingTimeout(ctx context.Context, state *State) (StateAction[State], error) {
	log.Println("Host timedout")
	state.currentPool = nil
	return SendRetrievePools, nil
}

func (s *State) Listen(to []shared.MessageType) (StateAction[State], error) {	
	var msgType byte 
	buff := make([]byte, 1024)
	for {
		nbr, err := s.conn.Read(buff)
		if err != nil {
			return nil, err
		}
		s.packet.Load(buff[:nbr])
		msgType, err = s.packet.ReadByte()
		if err != nil {
			return nil, err
		}
		if slices.Contains(to, shared.MessageType(msgType)) {
			break
		}
		log.Println("Read message", msgType)
	}	
	return s.messageHandlers[shared.MessageType(msgType)], nil
}
