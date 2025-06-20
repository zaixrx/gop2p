package broadcast

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/zaixrx/gop2p/shared"
	"github.com/zaixrx/gop2p/logging"
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

func (emt EMessageType) String() string {
	return [...]string{"JoinPool", "Create Pool", "Retreive Pools"}[emt]
}

type Handle struct {
	ctx context.Context
	cancel context.CancelFunc

	nm *NetworkManager

	emsg chan externalMessage
	eerr chan error
	
	Logger logging.Logger

	currentPool *shared.PublicPool
	poolIDs []string
}

func CreateHandle(pctx context.Context) *Handle {
	ctx, cancel := context.WithCancel(pctx)
	return &Handle{
		ctx: ctx,
		cancel: cancel,
		nm: NewNetworkManager(),
		emsg: make(chan externalMessage),
		eerr: make(chan error),
		Logger: logging.NewStdLogger(),
	}
}

func (h *Handle) Connect(hostname string, port uint16) error {
	err := h.nm.Connect(hostname, port)
	if err != nil {
		return err
	}
	h.Logger.Info("Connected to %s:%d\n", hostname, int(port))
	go h.handleUserRPCs()
	return nil
}

func (h *Handle) handleUserRPCs() error {
	var err error

	for {
		emsg := <-h.emsg

		select {
		case <-h.ctx.Done():
			h.Logger.Info("context done, stopping RPC handler...")
			h.Close()
			return nil 
		default:
			h.Logger.Debug("received RPC initiating call, for message of type %s...", emsg.msgType.String())
			// Validation
			switch emsg.msgType {
			case EMessageJoinPool:
				if len(emsg.args) != 1 {
					err = fmt.Errorf("expected poolID(string) got %d args", len(emsg.args))
					h.eerr <- err
					continue
				}

				// Send network call
				err = h.nm.SendJoinPool(emsg.args[0])
				if err != nil {
					h.eerr <- err
					continue
				}
				
				// Listen for response
				// Notice that this call always triggers an external error to resume execution
				// of the main goroutine
				h.eerr <- h.networkListen([]shared.MessageType{shared.MessageJoinPool})
			case EMessageCreatePool:
				if len(emsg.args) != 0 {
					h.eerr <- fmt.Errorf("expected no args got %d", len(emsg.args))
					continue
				}

				// Send network call
				err := h.nm.SendCreatePool()
				if err != nil {
					h.eerr <- err
					continue
				}

				// Listen for response
				h.eerr <- h.networkListen([]shared.MessageType{shared.MessageJoinPool})
			case EMessageRetrievePools:
				if len(emsg.args) != 0 {
					h.eerr <- fmt.Errorf("expected no args got %d", len(emsg.args))
					continue
				}

				// Send network call
				err := h.nm.SendRetrievePools()
				if err != nil {
					h.eerr <- err
					continue
				}

				// Listen for response
				h.eerr <- h.networkListen([]shared.MessageType{shared.MessageRetrievePools})
			default:
				h.eerr <- fmt.Errorf("unknown message type")
			}
		}	
	}
}

func (h *Handle) networkListen(toWhat []shared.MessageType) error {
	for {
		packet, err := h.nm.Listen()
		if err != nil {
			h.Logger.Error("closing connection with the broadcaster...")
			h.Close()
			return err
		}

		byt, err := packet.ReadByte()
		if err != nil {
			return err 
		}

		msgType := shared.MessageType(byt)
		if slices.Contains(toWhat, msgType) {
			h.Logger.Debug("received RPC response, for message of type %s", msgType.String())

			switch msgType {
			case shared.MessageRetrievePools:
				poolIDs, err := packet.ReadStringArr()
				if err != nil {
					return err
				}
				h.poolIDs = poolIDs
				h.Logger.Debug("set available pools")
				return nil
			case shared.MessageJoinPool:
				pool, err := packet.ReadPool()
				if err != nil {
					return err
				}
				h.currentPool = pool
				h.Logger.Debug("set the current pool")
				return nil
			case shared.MessageError:
				msg, err := packet.ReadString()
				if err != nil {
					return err 
				}
				return errors.New("broadcaster responds: " + msg) 
			}
		}
	}
}

func (h *Handle) Close() error {
	if h.nm == nil {
		return nil
	}

	h.Logger.Info("closed connection with the broadcaster")

	err := h.nm.Close()
	h.nm = nil

	close(h.eerr)
	close(h.emsg)

	h.cancel()

	return err
}

type PublicPool = shared.PublicPool

func (h *Handle) JoinPool(poolID string) (*PublicPool, error) {
	if h.nm == nil {
		return nil, fmt.Errorf("you must connect to a broadcaster first!")
	}

	h.emsg <- externalMessage{
		msgType: EMessageJoinPool,
		args: []string{poolID},
	}

	err := <-h.eerr
	if err != nil {
		return nil, err
	}

	// js.err gives a signal indicating the termination of the RPC thus it expects the job's state to contain
	// relevent data based on what the user demanded 
	return h.currentPool, nil
}

func (h *Handle) CreatePool() (*PublicPool, error) {
	if h.nm == nil {
		return nil, fmt.Errorf("you must connect to a broadcaster first!")
	}

	h.emsg <- externalMessage{
		msgType: EMessageCreatePool,
		args: []string{},
	}

	err := <-h.eerr
	if err != nil {
		return nil, err
	}

	return h.currentPool, nil
}

func (h *Handle) GetPoolIDs() ([]string, error) {
	if h.nm == nil {
		return nil, fmt.Errorf("you must connect to a broadcaster first!")
	}

	h.emsg <- externalMessage{
		msgType: EMessageRetrievePools,
		args: []string{},
	}

	err := <-h.eerr
	if err != nil {
		return nil, err
	}
	
	return h.poolIDs, nil	
}

func (h *Handle) Ping(pctx context.Context, ticks int) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	if h.nm == nil || h.currentPool == nil {
		return
	}

	if h.currentPool.HostIP != h.currentPool.YourIP {
		return
	}

	h.Logger.Info("started pinging the pool alive")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := h.networkListen([]shared.MessageType{shared.MessagePoolPingTimeout})
				if err != nil {
					continue
				}
				h.Logger.Info("pool timedout")
				cancel()
			}
		}	
	}()	

	limitter := time.Tick(time.Millisecond * time.Duration(1000 / ticks))
	for {
		select {
		case <-ctx.Done():
			// TODO: this can happen if the broadcaster shuts down without expection
			// or in a performance dropdown where client doesn't send ping message
			// so you must either throw and error, or reconnect to the broadcaster
			return
		case <-limitter:
			err := h.nm.SendPoolPingMessage(h.currentPool.Id)
			if err != nil {
				h.Logger.Error(err.Error())
			}
		}
	}
}
