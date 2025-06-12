package main

import (
	"log"
	"net"
	"p2p/shared"
	"strconv"
	"time"
)

func main() {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("Connected to broadcaster")

	go func() {
		pakt := shared.NewPacket(addr)
		pakt.WriteByte(byte(shared.GetPeersMessage))
		conn.Write(pakt.GetBytes())
		buff := make([]byte, 1024)
		nbr, err := conn.Read(buff)
		if err != nil {
			log.Fatal("1", err)
		}
		pakt.Load(buff[:nbr])
		str, err := pakt.ReadString()
		if err != nil {
			log.Fatal("2", err)
		}
		pakt.Flush()

		log.Println("Retreived peers", str)
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
		log.Println("hello")
	}
}
