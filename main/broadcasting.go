package main

import (
	"net"
	"log"
	"fmt"
	"time"
	"errors"
	"slices"
	"context"
	"strconv"
	"strings"
	"p2p/shared"
)

//type StateAction[State any] func(state *State) (StateAction[State], error)
type BRState struct {
	br *Broadcast

	conn net.Conn
	packet *shared.Packet
	listenHandler map[shared.MessageType]StateAction[BRState]

	Pools []string
	CurrentPool *shared.PublicPool
	
	onConnection func() (BRMsgType, []string) 
	closeChan chan struct{}
}

type BRMsgType byte 

const (
	BRCreatePool BRMsgType = iota 
	BRJoinPool
)

func CreateBroadcastingStage(closeChan chan struct{}, onConnection func() (BRMsgType, []string)) *StateContext[BRState] {
	stage := NewStateContext(ConnectToBroadcaster)
	
	stage.state.packet = shared.NewPacket()
	stage.state.listenHandler = map[shared.MessageType]StateAction[BRState]{
		shared.MessageRetrievePools: HandleRetreivePool,
		shared.MessageJoinPool: HandlePoolJoin,
	}

	stage.state.onConnection = onConnection
	stage.state.closeChan = closeChan
	
	return stage
}

func ConnectToBroadcaster(ctx context.Context, state *BRState) (StateAction[BRState], error) {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}

	state.conn = conn
	state.br = NewBroadcast(conn)	

	return SendRetrievePools, nil 
}

func SendRetrievePools(ctx context.Context, state *BRState) (StateAction[BRState], error) {
	err := state.br.SendPoolRetreivalMessage()
	if err != nil {
		return nil, err
	}

	next, err := state.Listen([]shared.MessageType{shared.MessageRetrievePools})
	if err != nil {
		// TODO: reconnect merda 
		return nil, err
	}

	return next, nil
}

func HandleRetreivePool(ctx context.Context, state *BRState) (StateAction[BRState], error) {
	rawPools, err := state.packet.ReadString()
	if err != nil {
		return nil, err
	}

	state.Pools = strings.Split(rawPools, " ") 
	log.Println(rawPools)
	// TODO: stream that shit over a channel

	return HandleAfterPoolRetrieval, nil 
}

func HandleAfterPoolRetrieval(ctx context.Context, state *BRState) (StateAction[BRState], error) {
	typ, args := state.onConnection()

	switch typ {
	case BRCreatePool:
		err := state.br.SendPoolCreateMessage()
		if err != nil {
			return HandleAfterPoolRetrieval, err 
		}
	case BRJoinPool:
		err := state.br.SendPoolJoinMessage(args[0])
		if err != nil {
			return HandleAfterPoolRetrieval, err 
		}
	}

	next, err := state.Listen([]shared.MessageType{shared.MessageJoinPool})
	if err != nil {
		return HandleAfterPoolRetrieval, err
	}

	return next, nil
}

func HandlePoolJoin(ctx context.Context, state *BRState) (StateAction[BRState], error) {
	pool, err := state.packet.ReadPool()
	if err != nil {
		return nil, err
	}

	state.CurrentPool = pool

	if state.CurrentPool.YourIP == state.CurrentPool.HostIP {
		go state.HandlePoolPing(pool.Id, state.closeChan)
		go state.HandlePingTimeout(state.closeChan)
	}

	return nil, nil
}

func (s *BRState) HandlePoolPing(poolID string, closeChan chan struct{}) {
	limitter := time.Tick(time.Millisecond * 1000 / shared.PoolPingTicks)
	for {
		select {
		case <-closeChan:
			return
		case <-limitter:
			err := s.br.SendPoolPingMessage(poolID)
			if err != nil {
				closeChan <- struct{}{}
				return
			}
		}
	}
}

func (s *BRState) HandlePingTimeout(closeChan chan<- struct{}) {
	s.Listen([]shared.MessageType{shared.MessagePoolPingTimeout})
	closeChan <- struct{}{}
}

// I do this to filter unexpected unsyncs with the server
// To assure that the read message is what I want
func (s *BRState) Listen(to []shared.MessageType) (StateAction[BRState], error) {	
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
		if msgType == byte(shared.MessageError) {
			errmsg, corrErrErr := s.packet.ReadString()
			if corrErrErr != nil {
				return nil, fmt.Errorf("ERROR: recieved unexpected error format from broadcaster %s", corrErrErr)
			}
			return nil, errors.New(errmsg)
		}
		if slices.Contains(to, shared.MessageType(msgType)) {
			break
		}
		log.Println("Read message", msgType)
	}	
	return s.listenHandler[shared.MessageType(msgType)], nil
}
