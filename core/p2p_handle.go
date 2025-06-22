package p2p

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zaixrx/gop2p/core/transport"
	"github.com/zaixrx/gop2p/logging"
	"github.com/zaixrx/gop2p/shared"
)


type Handle struct {
	ctx context.Context
	cancel context.CancelFunc
	nm *network_manager
	Logger logging.Logger
	onDisconnect func()
}

type PublicPool = shared.PublicPool
type Packet = shared.Packet
func NewPacket() *Packet {
	return shared.NewPacket()
}

func CreateHandle(pctx context.Context, onDisconnect func()) *Handle {
	ctx, cancel := context.WithCancel(pctx)
	return &Handle{
		ctx: ctx,
		cancel: cancel,
		onDisconnect: onDisconnect,
		nm: NewNetworkManager(),
		Logger: logging.NewStdLogger(),
	}
}

func (h *Handle) ConnectToPeer(addr string) (*Peer, error) {
	conn, err := h.nm.Connect(addr)
	if err != nil {
		return nil, err
	}
	peer := newPeer(h.ctx, &conn)
	return peer, nil 
}

func (h *Handle) ConnectToPool(pool *PublicPool) (map[string]*Peer, error) {
	peers := make(map[string]*Peer)
	builder := strings.Builder{}

	for _, addr := range pool.PeerIPs {
		if addr == pool.YourIP {
			continue
		}
		peer, err := h.ConnectToPeer(addr)
		if err != nil {
			builder.WriteString(err.Error() + "\n")
			continue
		}
		builder.WriteString("connection success\n")
		peers[peer.Addr] = peer
	}
	
	if builder.Len() > 0 {
		return peers, errors.New(builder.String())
	}

	return peers, nil
}

func (h *Handle) Listen(port uint16) error {
	err := h.nm.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	h.Logger.Info("Listening on port %d\n", port)
	return nil
}

func (h *Handle) Accept() (*Peer, error) {
	connChan := make(chan transport.Conn)
	errChan := make(chan error)

	go func() {
		conn, err := h.nm.Accept()
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	select {
	case <-h.ctx.Done():
		return nil, h.ctx.Err()
	case err := <- errChan:
		return nil, err
	case conn := <- connChan:
		p := newPeer(h.ctx, &conn)
		h.Logger.Debug("accepted new connection %s\n", p.Addr)
		return p, nil
	}
}
func (h *Handle) Close() error {
	h.cancel()	
	err := h.nm.Close()
	h.Logger.Info("closed peer listener\n")
	h.onDisconnect()
	return err
}

type Peer struct {
	Addr string

	ctx context.Context
	ctxCancel context.CancelFunc

	conn *transport.Conn
	
	handlersLock sync.Mutex
	handlers map[string]func(*transport.Packet)

	sendQLock sync.Mutex
	sendQ []*transport.Packet
}

const PeerDisconnected string = "disconnected"
var handlerZeroValue func(*transport.Packet) = func(_ *transport.Packet) {}
func newPeer(parentCtx context.Context, conn *transport.Conn) *Peer {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Peer{
		conn: conn,
		
		sendQ: make([]*transport.Packet, 0),
		handlers: map[string]func(*transport.Packet){
			PeerDisconnected: handlerZeroValue,
		},

		ctx: ctx,
		ctxCancel: cancel,

		Addr: (*conn).Address(),
	}
}

// Bridge between handle and peer
// Listens for new packets
func (h *Handle) HandlePeerIO(p *Peer) {
	conn := *p.conn

	go func() {
		limitter := time.Tick(time.Millisecond * time.Duration(1000 / 30))
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-limitter:
				for _, packet := range p.sendQ {
					msg, _ := packet.ReadString()
					packet.SetOffset(0)

					nbw, err := conn.Write(packet)
					if err != nil {
						h.Logger.Error("failed to send message to %s\n", p.Addr)
						continue
					}

					h.Logger.Debug("sent message %s of size %d to %s\n", msg, nbw, p.Addr)
				}
				p.sendQ = p.sendQ[:0] 
			}
		}

	}()

	go func() {
		limitter := time.Tick(time.Duration(1000 / 30) * time.Millisecond)
		for {
			select {
			case <-p.ctx.Done():
				p.handlers[PeerDisconnected](nil)
				return
			case <-limitter:
				packet, err := conn.Read()
				if err != nil {
					p.Disconnect()
					continue
				}

				msgType, err := packet.ReadString()
				if err != nil {
					continue // TODO: error
				}
				
				h.Logger.Debug("read message %s of size %d from %s\n", msgType, packet.GetLen(), p.Addr)

				handler, exists := p.handlers[msgType]
				if !exists {
					h.Logger.Warn("invalid message type, handler doesn't exist from %s", p.Addr)
					continue
				}

				handler(packet)
			}
		}
	}()
}

func (p *Peer) Send(msg string, packet *transport.Packet) error { // Need to register sent messages and handle them in one go
	packet.SetWriteBefore(true)
	packet.WriteString(msg)

	p.sendQLock.Lock()
	p.sendQ = append(p.sendQ, packet)
	p.sendQLock.Unlock()

	return nil
}

func (p *Peer) On(msg string, handler func(data *transport.Packet)) { // Needs read new packets from this peer
	p.handlersLock.Lock()
	p.handlers[msg] = handler
	p.handlersLock.Unlock()
}

func (p *Peer) Disconnect() error {
	p.ctxCancel()
	return nil
}

/*

Problems: 
		What should you really do when a peer from a pool fails to connect? reconnect
		Do retries if a connection doesn't want to be accepted, let the user do that

	  	Why the fuck aren't you batching sent packets and add a pipeline or somthing for e2e, out of context(besides the batching part which will need some sort of benchmarking)
*/

