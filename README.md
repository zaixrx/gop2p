# Example

```go
package main

import (
	"log"
  P2P "p2p/main/p2p"
	Broadcast "p2p/main/broadcast"
)

func main() {
	br := Broadcast.CreateBroadcast()
	go br.Start()

	pools, err := br.RetrievePools()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Available Pools", pools)

	currentPool, err := br.SendCreatePool()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created Pool:", currentPool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go br.Ping(ctx, cancel) // Used to keep the pool alive
	
	p2p := P2P.CreateP2P(pool)
	go p2p.Start(ctx, cancel)
	
	// "newConnection" is a built-in message type along with dozens of other messages
	p2p.On("newConnection", func(from string, _ *P2P.Packet) {
		p2p.Send("ping", from, nil)
	})
	
	p2p.On("ping", func(from string, _ *P2P.Packet) {
		log.Println("Recieved ping from %s")
	})
}
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
- A StateMachine model : https://medium.com/@johnsiilver/go-state-machine-patterns-3b667f345b5e
  https://www.youtube.com/watch?v=HxaD_trXwRE
- An Idea About Debouncing(Inspired me to check for pool timeouts): https://medium.com/gopher-time/implementing-debounce-functionality-in-go-29c4e7a83a56
- General Concepts: https://gobyexample.com/
