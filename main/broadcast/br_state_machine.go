package Broadcast

import (
	"fmt"
	"errors"
	"slices"
	"p2p/shared"
	machine "p2p/main/stateMachine"
)

type output struct {
	CurrentPool *shared.PublicPool
	Pools []string
}

type state struct {
	nm *NetworkManager
	packet *shared.Packet // Maybe make that in the nm

	emsg chan *externalMessage
	eerr chan error 
}

type externalMessage struct {
	msgType EMessageType
	args []string
}

type EMessageType byte
const (
	EMessageJoinPool EMessageType = iota
	EMessageCreatePool
	EMessageRetrievePools
)

type jobState struct {
	state *state
	output *output
}

type stateMachine machine.StateMachine[state, output]

func CreateBroadcast() *stateMachine {
	return (*stateMachine)(machine.NewStateMachine(state{
		nm: NewNetworkManager(shared.Hostname, shared.Port),
		emsg: make(chan *externalMessage),
		eerr: make(chan error),
	}, ConnectToBroadcaster))
}

func (sm *stateMachine) Start() {	
	(*machine.StateMachine[state, output])(sm).Run()
}

// This runs on a different goroutine as the state machine
func (sm *stateMachine) SendJoinPool(poolID string) (*shared.PublicPool, error) {
	rawSM := (*machine.StateMachine[state, output])(sm)
	state := rawSM.GetState() 

	state.emsg <- &externalMessage{
		msgType: EMessageJoinPool,
		args: []string{poolID},
	}
	err := <-state.eerr
	if err != nil {
		return nil, err
	}

	return rawSM.GetOutput().CurrentPool, nil	
}

func (sm *stateMachine) SendCreatePool() (*shared.PublicPool, error) {
	rawSM := (*machine.StateMachine[state, output])(sm)
	state := rawSM.GetState()

	state.emsg <- &externalMessage{
		msgType: EMessageCreatePool,
		args: []string{},
	}
	err := <-state.eerr
	if err != nil {
		return nil, err
	}

	return rawSM.GetOutput().CurrentPool, nil	
}

func (sm *stateMachine) RetrievePools() ([]string, error) {
	rawSM := (*machine.StateMachine[state, output])(sm)
	state := rawSM.GetState()
	
	state.emsg <- &externalMessage{
		msgType: EMessageRetrievePools,
		args: []string{},
	}
	err := <-state.eerr
	if err != nil {
		return nil, err
	}
	
	return rawSM.GetOutput().Pools, nil	
}

func ConnectToBroadcaster(s *state, o *output) (machine.StateJob[state, output], error) {
	err := s.nm.Connect()
	if err != nil {
		// Terminating error 
		return nil, err 
	}
	return HandleRetrievePools, nil 
}

func HandleRetrievePools(s *state, o *output) (machine.StateJob[state, output], error) {	
	js := &jobState{
		state: s,
		output: o,
	}
	return js.UserListen([]EMessageType{EMessageRetrievePools})
}

func HandleJoinPool(s *state, o *output)(machine.StateJob[state, output], error) {
	js := &jobState{
		state: s,
		output: o,
	}
	return js.UserListen([]EMessageType{EMessageCreatePool, EMessageJoinPool})
}

func (js *jobState) UserListen(toWhat []EMessageType)(machine.StateJob[state, output], error) {
	for {
		emsg := <-js.state.emsg

		// Validation
		if !slices.Contains(toWhat, emsg.msgType) {
			js.state.eerr <- fmt.Errorf("ERROR: invalid message type")
		}

		switch emsg.msgType {
		case EMessageJoinPool:
			if len(emsg.args) != 1 {
				js.state.eerr <- fmt.Errorf("ERROR: expected poolID(string) got %d args", len(emsg.args))
			}

			// Send network call
			err := js.state.nm.SendJoinPool(emsg.args[0])
			if err != nil {
				js.state.eerr <- err
			}
			
			// Listen for response
			next, err := js.NetworkListen([]shared.MessageType{shared.MessageJoinPool})
			js.state.eerr <- err

			return next, nil
		case EMessageCreatePool:
			if len(emsg.args) != 0 {
				js.state.eerr <- fmt.Errorf("ERROR: expected no args got %d", len(emsg.args))
			}

			// Send network call
			err := js.state.nm.SendCreatePool()
			if err != nil {
				js.state.eerr <- err
			}
			// Listen for response
			next, err := js.NetworkListen([]shared.MessageType{shared.MessageJoinPool})
			js.state.eerr <- err

			return next, nil
		case EMessageRetrievePools:
			if len(emsg.args) != 0 {
				js.state.eerr <- fmt.Errorf("ERROR: expected no args got %d", len(emsg.args))
			}

			// Send network call
			err := js.state.nm.SendRetrievePools()
			if err != nil {
				js.state.eerr <- err
			}

			// Listen for response
			next, err := js.NetworkListen([]shared.MessageType{shared.MessageRetrievePools})
			js.state.eerr <- err

			return next, nil
		default:
			js.state.eerr <- fmt.Errorf("ERROR: unknown message type")
		}
	}
}

func (js *jobState) NetworkListen(toWhat []shared.MessageType)(machine.StateJob[state, output], error) {
	for {
		packet, err := js.state.nm.Listen()
		if err != nil {
			return nil, err
		}
		byt, err := packet.ReadByte()
		if err != nil {
			return nil, err
		}
		msgType := shared.MessageType(byt)
		if slices.Contains(toWhat, msgType) {
			switch msgType {
			case shared.MessageRetrievePools:
				pools, err := packet.ReadStringArr()
				if err != nil {
					return nil, err
				}
				js.output.Pools = pools
				return HandleJoinPool, nil
			case shared.MessageJoinPool:
				pool, err := packet.ReadPool()
				if err != nil {
					return nil, err
				}
				js.output.CurrentPool = pool
				return nil, nil
			case shared.MessageError:
				msg, err := packet.ReadString()
				if err != nil {
					return nil, err
				}
				return nil, errors.New(msg)
			}
		}
	}
}
