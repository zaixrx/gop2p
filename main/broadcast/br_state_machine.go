package Broadcast

import (
	"fmt"
	"time"
	"errors"
	"slices"
	"context"
	"p2p/shared"
	machine "p2p/main/stateMachine"
)

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

type t_job_state struct {
	nm *NetworkManager

	emsg chan *externalMessage
	eerr chan error

	pools []string
	currentPool *shared.PublicPool
}

type stateMachine machine.StateMachine[t_job_state]

func CreateBroadcast(hostName string, port uint16) *stateMachine {
	return (*stateMachine)(machine.NewStateMachine(t_job_state{
		nm: NewNetworkManager(hostName, port),
		emsg: make(chan *externalMessage),
		eerr: make(chan error),
	}, ConnectToBroadcaster))
}

func ConnectToBroadcaster(ctx context.Context, js *t_job_state) (machine.StateJob[t_job_state], error) {
	err := js.nm.Connect()
	if err != nil {
		// Terminating error 
		return nil, err 
	}
	return HandleRetrievePools, nil 
}

func HandleRetrievePools(ctx context.Context, js *t_job_state) (machine.StateJob[t_job_state], error) {	
	for {
		next, err := js.UserListen([]EMessageType{EMessageRetrievePools})
		if err != nil {
			return HandleRetrievePools, err	
		}
		return next, nil
	}
}

func HandleJoinPool(ctx context.Context, js *t_job_state)(machine.StateJob[t_job_state], error) {
	for {
		next, err := js.UserListen([]EMessageType{EMessageCreatePool, EMessageJoinPool})
		if err != nil {
			return HandleJoinPool, err	
		}
		return next, nil
	}
}

func HandleChannelDisposal(ctx context.Context, js *t_job_state)(machine.StateJob[t_job_state], error) {
	close(js.eerr)
	close(js.emsg)
	return nil, nil
}

func (js *t_job_state) UserListen(toWhat []EMessageType)(machine.StateJob[t_job_state], error) {
	var err error

	for {
		emsg := <-js.emsg

		// Validation
		if !slices.Contains(toWhat, emsg.msgType) {
			err = fmt.Errorf("ERROR: unvalid operation at the current state, valid op codes are : %v", toWhat)
			js.eerr <- err
			return nil, err
		}

		switch emsg.msgType {
		case EMessageJoinPool:
			if len(emsg.args) != 1 {
				err = fmt.Errorf("ERROR: expected poolID(string) got %d args", len(emsg.args))
				js.eerr <- err
				return nil, err
			}

			// Send network call
			err = js.nm.SendJoinPool(emsg.args[0])
			if err != nil {
				js.eerr <- err
				return nil, err
			}
			
			// Listen for response
			next, err := js.NetworkListen([]shared.MessageType{shared.MessageJoinPool})
			// Notice that this call allways triggers the external error channel to resume execution
			// of the main goroutine
			js.eerr <- err
			if err != nil {
				return nil, err
			}

			return next, nil
		case EMessageCreatePool:
			if len(emsg.args) != 0 {
				js.eerr <- fmt.Errorf("ERROR: expected no args got %d", len(emsg.args))
			}

			// Send network call
			err := js.nm.SendCreatePool()
			if err != nil {
				js.eerr <- err
			}
			// Listen for response
			next, err := js.NetworkListen([]shared.MessageType{shared.MessageJoinPool})
			js.eerr <- err

			return next, nil
		case EMessageRetrievePools:
			if len(emsg.args) != 0 {
				js.eerr <- fmt.Errorf("ERROR: expected no args got %d", len(emsg.args))
			}

			// Send network call
			err := js.nm.SendRetrievePools()
			if err != nil {
				js.eerr <- err
			}

			// Listen for response
			next, err := js.NetworkListen([]shared.MessageType{shared.MessageRetrievePools})
			js.eerr <- err

			return next, nil
		default:
			js.eerr <- fmt.Errorf("ERROR: unknown message type")
		}
	}
}

func (js *t_job_state) NetworkListen(toWhat []shared.MessageType)(machine.StateJob[t_job_state], error) {
	for {
		packet, err := js.nm.Listen()
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

				js.pools = pools
				return HandleJoinPool, nil
			case shared.MessageJoinPool:
				pool, err := packet.ReadPool()
				if err != nil {
					return nil, err
				}

				js.currentPool = pool
				return HandleChannelDisposal, nil
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

func (sm *stateMachine) Start(ctx context.Context) {	
	(*machine.StateMachine[t_job_state])(sm).Run(ctx)
}

// This runs on a different goroutine as the state machine
func (sm *stateMachine) JoinPool(poolID string) (*shared.PublicPool, error) {
	rawSM := (*machine.StateMachine[t_job_state])(sm)
	js := rawSM.GetState() 

	js.emsg <- &externalMessage{
		msgType: EMessageJoinPool,
		args: []string{poolID},
	}

	err := <-js.eerr
	if err != nil {
		return nil, err
	}

	// js.err gives a signal indicating the termination of the RPC thus it expects the job's state to contain
	// relevent data based on what we oredered
	return rawSM.GetState().currentPool, nil
}

func (sm *stateMachine) CreatePool() (*shared.PublicPool, error) {
	rawSM := (*machine.StateMachine[t_job_state])(sm)
	js := rawSM.GetState()

	js.emsg <- &externalMessage{
		msgType: EMessageCreatePool,
		args: []string{},
	}

	err := <-js.eerr
	if err != nil {
		return nil, err
	}

	return rawSM.GetState().currentPool, nil
}

func (sm *stateMachine) RetreivePools() ([]string, error) {
	rawSM := (*machine.StateMachine[t_job_state])(sm)
	js := rawSM.GetState()

	js.emsg <- &externalMessage{
		msgType: EMessageRetrievePools,
		args: []string{},
	}

	err := <-js.eerr
	if err != nil {
		return nil, err
	}
	
	return rawSM.GetState().pools, nil	
}

const Ticks int = 10
func (sm *stateMachine) Ping(ctx context.Context) {
	rawSM := (*machine.StateMachine[t_job_state])(sm)
	js := rawSM.GetState()

	if js.nm == nil || js.currentPool == nil {
		return
	}

	limitter := time.Tick(time.Millisecond * time.Duration(1000 / Ticks))
	for {
		select {
		case <-ctx.Done():
			// TODO: this can happen if the broadcaster shuts down without expection
			// or in a performance dropdown where client doesn't send ping message
			// so you must either throw and error, or reconnect to the broadcaster
			fmt.Println("Closed Ping")
			return
		case <-limitter:
			js.nm.SendPoolPingMessage(js.currentPool.Id)
		}
	}
}

func (sm *stateMachine) Stop() error {
	rawSM := (*machine.StateMachine[t_job_state])(sm)
	js := rawSM.GetState()

	err := js.nm.Close()
	js.nm = nil
	
	return err 
}
