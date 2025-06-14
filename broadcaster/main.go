package main

import (
	"log"
	"net"
	"p2p/shared"
)

var server *Server

func main() {
	addr := net.UDPAddr{
		IP: net.ParseIP(shared.Hostname),
		Port: shared.Port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Coulnd't listen on %d: %s\n", shared.Port, err) 
	}	

	defer conn.Close()
	log.Printf("Listening on %s:%d\n", shared.Hostname, shared.Port)

	server = NewServer(conn)
	
	for {
		err := server.Listen()
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			continue
		}
	}
}
