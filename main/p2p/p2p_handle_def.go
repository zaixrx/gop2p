package P2P

import (
	"fmt"
	"time"
	"sync"
	"errors"
	"strings"
	"context"
	"p2p/shared"
)

type Handle struct {
	ctx context.Context
	cancel context.CancelFunc

	nm *network_manager
}

func CreateHandle() *Handle {
	ctx, cancel := context.WithCancel(context.Background())
	return &Handle{
		ctx: ctx,
		cancel: cancel,
		nm: NewNetworkManager(),
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

func (h *Handle) ConnectToPool(pool *shared.PublicPool) (map[string]*Peer, error) {
	peers := make(map[string]*Peer)
	builder := strings.Builder{}

	for _, addr := range pool.PeerIPs {
		if addr == pool.YourIP {
			continue
		}
		peer, err := h.ConnectToPeer(addr)
		if err != nil {
			builder.WriteString(err.Error())
			continue
		}
		peers[peer.Addr] = peer
	}
	
	if builder.Len() > 0 {
		return peers, errors.New(builder.String())
	}

	return peers, nil
}

func (h *Handle) Listen(port uint16) error {
	return h.nm.Listen(fmt.Sprintf(":%d", port))
}

func (h *Handle) Accept() (*Peer, error) {
	if !h.nm.listening {
		return nil, fmt.Errorf("You must listen before accepting new peers, dumb ass!")
	}
	conn, err := h.nm.Accept()
	if err != nil {
		return nil, err
	}
	p := newPeer(h.ctx, &conn)
	return p, nil
}

func (h *Handle) Close() error {
	h.cancel()	
	err := h.nm.Close()
	h.nm = nil
	return err
}

type Peer struct {
	Addr string

	ctx context.Context
	ctxCancel context.CancelFunc

	conn *t_conn
	
	handlersLock sync.Mutex
	handlers map[string]func(*Packet)

	sendQLock sync.Mutex
	sendQ []*Packet
}

const DisconnectedMessage string = "disconnected"
var handlerZeroValue func(*Packet) = func(_ *Packet) {}

func newPeer(parentCtx context.Context, conn *t_conn) *Peer {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Peer{
		conn: conn,
		
		sendQ: make([]*Packet, 0),
		handlers: make(map[string]func(*Packet)),

		ctx: ctx,
		ctxCancel: cancel,

		Addr: (*conn).Address(),
	}
}

// Bridge between handle and peer
// Listens for new packets
func (h *Handle) HandlePeer(p *Peer) {
	conn := *p.conn

	go func() {
		limitter := time.Tick(time.Millisecond * time.Duration(1000 / 30))
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-limitter:
				for _, packet := range p.sendQ {
					_, err := conn.Write(packet)
					if err != nil {
						fmt.Println("Failed to send message")
						continue
					}
					fmt.Println("Sent message")
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
				p.handlersLock.Lock()
				p.handlers[DisconnectedMessage](nil)
				p.handlersLock.Unlock()
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

				fmt.Println("Read result message")

				handler, exists := p.handlers[msgType]
				if !exists {
					fmt.Println("ERROR: invalid message type, handler doesn't exist")
					continue
				}

				handler(packet)
			}
		}
	}()
}

func (p *Peer) Send(msg string, packet *Packet) error { // Need to register sent messages and handle them in one go
	packet.SetWriteBefore(true)
	packet.WriteString(msg)
	packet.SetWriteBefore(false)
	p.sendQLock.Lock()
	p.sendQ = append(p.sendQ, packet)
	p.sendQLock.Unlock()
	return nil
}
func (p *Peer) On(msg string, handler func(data *Packet)) { // Needs read new packets from this peer
	fmt.Println("Set", msg)
	p.handlersLock.Lock()
	p.handlers[msg] = handler
	p.handlersLock.Unlock()
}
func (p *Peer) Disconnect() error {
	p.ctxCancel()
	return nil
}

/*

Problems: What should you really do when a peer from a pool fails to connect?
	  Do retries if a connection doesn't want to be accepted
	  Why the fuck aren't you batching sent packets and add a pipeline or somthing for e2e
*/

