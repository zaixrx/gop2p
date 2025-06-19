You want docs? read the fucking source code!

# Disclamer
I wrote this library for personal use, as a lot of my projects need it in one way or another, which means that I will still update this as I'm making my other project inshaa'allah, you are encouraged to contribute! just don't change the core systems

# Example

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	br := Broadcast.CreateBroadcast(BRHostname, BRPort)
	go br.Start(ctx)

	defer func () {
		br.Stop()
		cancel()
	}()

	pools, err := br.RetreivePools()
	if err != nil {
		log.Fatal(err)
	}

	pool, err = br.JoinPool(pools[0])
	if err != nil {
        	log.Fatal(err)
	}
    
    	if peer.HostIP == peer.YourIP {
        	go br.Ping(ctx)
    	}
	
	handle := P2P.CreateHandle()
	peers, err := peer.ConnectToPool(pool, handlePeer)

    	handlePeer = func(paddr string) {
	        peer := peers[paddr]
	
	        peer.On("msg", func(packet *P2P.Packet) {
	            msg, err := packet.ReadString()
	            log.Println(msg)
	        }

		packet := P2P.NewPacket()
		packet.WriteString("Hello, World!\r\n")
	        peer.Send("msg", packet)

        	handle.HandlePeerIO(peer)
	}

    	for _, p := range peers {
        	handlePeer(p)
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

# Source Code

br from p2p/main/broadcast:

implements a broadcasting manager that has a set of methods that acts as RPCs
```go
br.RetrievePools() ([]string, error)
br.SendCreatePool() (*PublicPool, error)
br.SendJoinPool(poolID string) (*PublicPool, error)
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
