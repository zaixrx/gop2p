// This provides a comprehensive example of this implementation
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	broadcast "github.com/zaixrx/gop2p/core/broadcast"
	p2p "github.com/zaixrx/gop2p/core/p2p"
	"github.com/zaixrx/gop2p/shared"
)

const (
	BRHostname string = "127.0.0.1"
	BRPort     uint16 = 6969
	BRPingTicks int = 10
)

func main() {
	var (
		err error
		poolIDs []string = []string{}
		pool *shared.PublicPool
	)

	ctx, cancel := context.WithCancel(context.Background())
	br := broadcast.CreateHandle(ctx)

	if err = br.Connect(BRHostname, BRPort); err != nil {
		log.Panic(err)
	}

	defer func () {
		br.Close()
		cancel()
	}()

	func () {
		for {
			c, _ := cmdSelect(map[string]int{
				"list": 0,
				"create": 0,
				"join": 0,
			})

			switch c {
			case "create":
				pool, err = br.CreatePool()
				if err != nil {
					log.Printf("ERROR: %s\n", err)
					continue
				}
				return
			case "join":
				if len(poolIDs) == 0 {
					log.Printf("ERROR: no pools to join\n")
					continue
				}

				pool, err = br.JoinPool(poolIDs[0])
				if err != nil {
					log.Printf("ERROR: invalid join pool %s\n", err)
					continue
				}

				return
			case "list":
				poolIDs, err = br.GetPoolIDs()
				if err != nil {
					log.Printf("ERROR: couldn't list pools %s\n", err)
					continue
				}
				log.Println(poolIDs)
			}
		}
	}()

	log.Println("Joined Pool With Peers", pool.PeerIPs)

	if pool.HostIP == pool.YourIP {
		go br.Ping(ctx, BRPingTicks)
	}

	/////////////////////////////////////////////////////////////////////////////

	handle := p2p.CreateHandle()
	peers, _ := handle.ConnectToPool(pool)

	handlePeer := func(paddr string) {
		p, exists := peers[paddr]
		if !exists {
			return
		}

		p.On("msg", func(p *p2p.Packet) {
			r, err := p.ReadString()
			if err != nil {
				return
			}
			log.Println(r)
		})
		p.On(p2p.DisconnectedMessage, func(_ *p2p.Packet) {
			delete(peers, p.Addr)

			if p.Addr == pool.HostIP {
				log.Println("Host left!")
				handle.Close()
			}
		})

		handle.HandlePeerIO(p)
	}

	for _, p := range peers {
		handlePeer(p.Addr)
	}
	
	port, err := extractPort(pool.YourIP)
	if err != nil {
		log.Fatal(err)
	}

	err = handle.Listen(port)
	if err != nil {
		log.Panic(err)
	}

	for {
		p, err := handle.Accept()
		if err != nil {
			log.Printf("ERROR: couldn't accept connection %s", err.Error())
		}

		peers[p.Addr] = p
		handlePeer(p.Addr)

		packet := p2p.NewPacket()
		packet.WriteString(fmt.Sprintf("I got your ip bitch! haha %s", p.Addr))

		p.Send("msg", packet)
	}
}

func cmdSelect(opts map[string]int) (string, []string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		log.Printf("Enter value from %v", opts)
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

func extractPort(addr string) (uint16, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	return uint16(port), err
}
