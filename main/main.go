package main

import (
	"log"
	"net"
	"p2p/shared"
	"strconv"
)

func main() {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	pakt := shared.NewPacket(addr)
	pakt.WriteByte(byte(shared.GetPeers))

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

	log.Println(str)
}
