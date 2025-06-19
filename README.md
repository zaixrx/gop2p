You want docs? read the fucking source code dumb ass!

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
	
	go br.Ping(ctx)
	
	peer := P2P.CreatePeer(Port)
	peers, err := peer.ConnectToPool(pool) 
	
	go func() {
		for {
		    p, err := peer.Accept()
		    if err != nil {
			continue
		    }
		    peers = append(peers, peer) 
		}
	}()

	peer.On("msg", func(from *P2P.Peer, pack *P2P.Packet) {
		msg, err := pack.ReadString()
		if err != nil {
			log.Printf("Received invalid message from %s", from)
			return
		}
		log.Println(msg)
	})
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

P2P Computing:

- General Idea: https://medium.com/@asafkozovsky/what-are-p2p-networks-exactly-77284fe3b8a3
  https://www.geeksforgeeks.org/system-design/peer-to-peer-p2p-architecture/
  https://en.wikipedia.org/wiki/Peer-to-peer#See_also
- Bit torrent(Systems were highly Inspired by bittorrent): https://web.cs.ucla.edu/classes/cs217/05BitTorrent.pdf
- An Example(golang): https://dev.to/hadeedtariq/hands-on-with-p2p-networks-building-a-messaging-system-12nd

Golang
- Concurrent Programming: https://edu.anarcho-copy.org/Programming%20Languages/Go/Concurrency%20in%20Go.pdf
- Concurrent Programming: https://edu.anarcho-copy.org/Programming%20Languages/Go/Concurrency%20in%20Go.pdf
- A StateMachine model : https://medium.com/@johnsiilver/go-state-machine-patterns-3b667f345b5e
- Concurrent Programming: https://edu.anarcho-copy.org/Programming%20Languages/Go/Concurrency%20in%20Go.pdf
  https://www.youtube.com/watch?v=HxaD_trXwRE
- An Idea About Debouncing(Inspired me to check for pool timeouts): https://medium.com/gopher-time/implementing-debounce-functionality-in-go-29c4e7a83a56
- General Concepts: https://gobyexample.com
