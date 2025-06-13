package main

import (
	"p2p/shared"
	"strings"
	"strconv"
	"bufio"
	"log"
	"net"
	"os"
)

type ActionHandler func() error

type AppState struct {
	pools []string
	currentPool string
	action map[string]ActionHandler
}

func main() {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Panic(err)
	}

	broadcast := NewBroadcast(conn)
	broadcast.SendPoolRetreivalMessage()

	log.Printf("Connected to %s\n", conn.RemoteAddr().String())

	var state AppState = AppState{
		action: map[string]ActionHandler{
			"create": broadcast.SendPoolCreateMessage,
			"join": func () error {
				reader := bufio.NewReader(os.Stdin)
				log.Print("Please enter the pool's id: ")
				id, err := reader.ReadString('\n') 
				if err != nil {
					return err
				}
				return broadcast.SendPoolJoinMessage(id[:len(id)-1])
			}, 
		},
	}

	for {
		packet, err := broadcast.Listen()
		if err != nil {
			log.Panic(err)
		}

		typ, err := packet.ReadByte()
		if err != nil {
			log.Panic(err)
		}

		log.Printf("Received message %d\n", typ)

		if state.pools == nil {
			if typ == byte(shared.MessageRetreivePools) {
				str, _ := packet.ReadString()
				state.pools = strings.Split(str, " ") 

				log.Println(state.pools)
			}
			go func() {
				for {
					var cmd string
					reader := bufio.NewReader(os.Stdin)
				readcmd:
					cmd, err = reader.ReadString('\n')
					if err != nil {
						log.Println(err)
						goto readcmd	
					}
					cmd = cmd[:len(cmd)-1]
					handler, ok := state.action[cmd]
					if !ok {
						log.Printf("SYNTAX_ERROR: Invalid command %s\n", cmd)
						goto readcmd
					}
					handler()
				}
			}()
			continue
		}

		if typ == byte(shared.MessageCreatePool) {
			str, _ := packet.ReadString()
			state.pools = append(state.pools, str)

			log.Println(state.pools)
		}
	}
}
