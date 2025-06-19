package main

import (
	"os"
	"log"
	"bufio"
	"context"
	"strings"
	"strconv"
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

	c, _ := cmdSelect(map[string]int{
		"create": 0,
		"join": 0,
	})

	var pool *shared.PublicPool
	switch c {
	case "create":
		pool, err = br.CreatePool()
		if err != nil {
			log.Fatal("CREATE_POOL:",err)
		}
	case "join":
		pool, err = br.JoinPool(pools[0])
		if err != nil {
			log.Fatal("JOIN Pool", err)
		}
	}

	log.Println("Joined Pool With Peers", pool.PeerIPs)

	if pool.HostIP == pool.YourIP {
		go br.Ping(ctx)
	}

	/////////////////////////////////////////////////////////////////////////////

	handle := P2P.CreateHandle()
	peers, _ := handle.ConnectToPool(pool)

	handlePeer := func(paddr string) {
		p, exists := peers[paddr]
		if !exists {
			return
		}

		p.On("msg", func(p *P2P.Packet) {
			r, _ := p.ReadString()
			log.Println(r)
		})
		p.On(P2P.DisconnectedMessage, func(_ *P2P.Packet) {
			log.Println("Peer disconnected")
			delete(peers, p.Addr)

			if p.Addr == pool.HostIP {
				log.Println("Host left!")
				handle.Close()
			}
		})
		handle.HandlePeer(p)
	}

	for _, p := range peers {
		handlePeer(p.Addr)
	}

	go func() {
		for {
			_, args := cmdSelect(map[string]int{
				"msg": 1,
	 		})
			for addr, p := range peers {
				if addr == pool.YourIP {
	 				continue
	 			}

				packet := shared.NewPacket()
	 			packet.WriteString(args[0])
	 			
	 			log.Println("Intending to send a message")
	 			p.Send("msg", packet)
	 		}
	 	}
	 }()

	port, err := strconv.Atoi(strings.Split(pool.YourIP, ":")[1])
	err = handle.Listen(uint16(port))
	if err != nil {
		log.Panic(err)
	}
	log.Println("Listening on port", port) 
	for {
		p, _ := handle.Accept()
		log.Println("New peer!")
		peers[p.Addr] = p
		handlePeer(p.Addr)
	}
}

func cmdSelect(opts map[string]int) (string, []string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			continue		
		}
		cmd = cmd[:len(cmd)-1]

		n, exists := opts[cmd]
		if !exists {
			log.Println("ERROR: unknown command")
			continue
		}

		args := make([]string, n)

		for i := range n {
			arg, _ := reader.ReadString('\n')
			args[i] = arg[:len(arg)-1]
		}
		
		return cmd, args
	}
}
