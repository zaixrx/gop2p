package main

import (
	"log"
	"net"
	"p2p/shared"
	"strconv"
	"time"
	"strings"
	"flag"
	"os"
	"bufio"
)

func main() {
	port := flag.Uint("port", 8080, "Peer tcp port")
	flag.Parse()

	peersChan := make(chan string)	
	go broadcast(peersChan)

	peers := strings.Split(<-peersChan, " ")
	
	_, err := NewPeer(uint16(*port), peers)
	if err != nil {
		log.Fatal(err)
	}
	
	// stdReader := bufio.NewReader(os.Stdin)
	// for {
	// 	peer.Read()	
	// }
}

func broadcast(peers chan string) {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		conn.Close()
		log.Println("Disconnected.")
	}()

	log.Println("Connected to broadcaster")

	go func() {
		pakt := shared.NewPacket(addr)
		pakt.WriteByte(byte(shared.GetPeersMessage))
		conn.Write(pakt.GetBytes())
		buff := make([]byte, 1024)
		nbr, err := conn.Read(buff)
		if err != nil {
			log.Fatal(err)
		}
		pakt.Load(buff[:nbr])
		str, err := pakt.ReadString()
		if err != nil {
			log.Fatal(err)
		}
		pakt.Flush()
		peers <- str
	}()

	pakt := shared.NewPacket(addr)
	pakt.WriteByte(byte(shared.PingMessage))
	dat := pakt.GetBytes()
	lim := time.Tick(time.Second / shared.PingTicks)
	for {
		<-lim
		nbr, err := conn.Write(dat)
		if err != nil || nbr < len(dat) {
			log.Printf("Failed to send message: %s\n", err)
			break
		}
	}
}
