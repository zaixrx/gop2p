# Disclamer
I wrote this library for personal use, as a lot of my projects need it in one way or another, which means that I will still update this as I'm making my other project inshaa'allah, you are encouraged to contribute! just don't change the core systems

# Example

Main Program (see [https://github.com/zaixrx/gop2p/blob/main/core/examples/main.go]):
```go
package bytetorrent

import (
	"log"
	"context"
	"github.com/zaixrx/gop2p/core/broadcast"
	P2P "github.com/zaixrx/gop2p/core/p2p"
)

const (
	BRPingTicks = 20
	BRHostname = "localhost"
	BRPort = 6969
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	br := broadcast.CreateHandle(ctx)
	if err := br.Connect(BRHostname, BRPort); err != nil {
		log.Panic(err)
	}

	defer func () {
		br.Close()
		cancel()
	}()

	pools, err := br.GetPoolIDs()
	if err != nil  {
		log.Fatal(err)
	}

	log.Println(pools)

	pool, err := br.CreatePool()
	if err != nil {
        	log.Fatal(err)
	}

	log.Println("Joined Pool With Peers", pool.PeerIPs)

    	if pool.HostIP == pool.YourIP {
		go br.Ping(ctx, BRPingTicks)
	}
	
	handle := P2P.CreateHandle()
	peers, err := handle.ConnectToPool(pool)

	if err != nil {
		handle.Close()
		log.Fatal(err)
	}

	handlePeer := func(paddr string) {
	        peer := peers[paddr]
	
	        peer.On("msg", func(packet *P2P.Packet) {
	            	msg, err := packet.ReadString()
                    	if err != nil {
				return
			}
	            	log.Println(msg)
	        })

		packet := P2P.NewPacket()
		packet.WriteString("Hello, World!\r\n")

	        peer.Send("msg", packet)

        	handle.HandlePeerIO(peer)
	}

    	for _, p := range peers {
        	handlePeer(p.Addr)
    	}

	for {
        	peer, err := handle.Accept()
        	if err != nil {
            		continue
        	}
        	peers[peer.Addr] = peer
        	handlePeer(peer.Addr)
    	}
}
```

Broadcasting Server (see[https://github.com/zaixrx/gop2p/blob/main/broadcaster/examples/main.go]):
```go
var server *broadcaster.Server

const (
	Port int = 6969
	PoolPingTimeout int = 1
	Hostname string = "127.0.0.1"
)

func main() {
	server = broadcaster.NewServer(PoolPingTimeout)
	server.Start(Hostname, Port)
	for {
		err := server.Listen()
		if err != nil {
			continue
		}
	}
}
```

# Source Code

br from p2p/main/broadcast:

broadcasting handle:
```go
func (h *Handle) Connect(hostname string, port uint16) error
func (h *Handle) Close() error

// These are RPCs
func (h *Handle) JoinPool(poolID string) (*PublicPool, error)
func (h *Handle) CreatePool() (*PublicPool, error)
func (h *Handle) GetPoolIDs() ([]string, error)

// Must be called to keep the pool alive
func (h *Handle) Ping(pctx context.Context, ticks int)
```

p2p handle:
```go
func (h *Handle) ConnectToPeer(addr string) (*Peer, error)
func (h *Handle) ConnectToPool(pool *shared.PublicPool) (map[string]*Peer, error)
func (h *Handle) Listen(port uint16) error
func (h *Handle) Accept() (*Peer, error)
func (h *Handle) Close() error

// Must be called after connecting or accepting a new peer
func (h *Handle) HandlePeerIO(p *Peer)

func (p *Peer) Send(msg string, packet *Packet) error
func (p *Peer) On(msg string, handler func(data *Packet))
func (p *Peer) Disconnect() error
```

# Resources

- p2p general idea: https://medium.com/@asafkozovsky/what-are-p2p-networks-exactly-77284fe3b8a3
                    https://www.geeksforgeeks.org/system-design/peer-to-peer-p2p-architecture/
                    https://en.wikipedia.org/wiki/Peer-to-peer#See_also
- an overview of how bit torrent manages p2p: https://web.cs.ucla.edu/classes/cs217/05BitTorrent.pdf
- how stream based sockets are handled kernel level: https://dev.to/hgsgtk/how-go-handles-network-and-system-calls-when-tcp-server-1nbd
- concurrency & parallelism: https://edu.anarcho-copy.org/Programming%20Languages/Go/Concurrency%20in%20Go.pdf
- an idea about state machines: https://medium.com/@johnsiilver/go-state-machine-patterns-3b667f345b5e
                                https://www.youtube.com/watch?v=HxaD_trXwRE
- an idea about debouncing: https://medium.com/gopher-time/implementing-debounce-functionality-in-go-29c4e7a83a56
- golang: https://gobyexample.com
