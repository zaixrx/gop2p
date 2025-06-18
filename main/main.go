package main

import (
	"os"
	"log"
	"bufio"
	"context"
	"p2p/shared"
	"p2p/main/p2p"
	"p2p/main/broadcast"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	br := Broadcast.CreateBroadcast(shared.Hostname, shared.Port)
	go br.Start(ctx)

	defer func () {
		br.Stop()
		cancel()
	}()

	pools, err := br.RetreivePools()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Available Pools", pools)

	var pool *shared.PublicPool
	reader := bufio.NewReader(os.Stdin)
	stop := false
	for !stop {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		cmd = cmd[:len(cmd)-1]

		switch cmd {
		case "create":
			pool, err = br.CreatePool()
			if err != nil {
				log.Fatal(err)
			}
			stop = true
		case "join":
			if len(pools) == 0 {
				log.Println("Can't join with no available pools")
				continue
			}
			pool, err = br.JoinPool(pools[0])
			if err != nil {
				log.Fatal(err)
			}
			stop = true	
		default:
			log.Printf("Invalid Command: received %s", cmd)
		}
	}

	go br.Ping(ctx)

	p2p := P2P.CreateP2P(pool, 0) // 0 for random port generation (based on your configured transport)
	go p2p.Run(ctx)


	// p2pConn, err := p2p.Connect(pool)
	// p2pConn.Accept()
	// p2p.Listen
	// p2pConn.Send

	p2p.On("msg", func(from string, pack *P2P.P2PPacket) {
		msg, err := pack.ReadString()
		if err != nil {
			log.Printf("Received invalid message from %s", from)
			return
		}
		log.Println(msg)
	})

	bufio.NewReader(os.Stdin).ReadString('\n')
}	
