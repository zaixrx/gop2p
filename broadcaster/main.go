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
		peer, err := server.Read()
		if err != nil {
			log.Printf("ERROR: couldn't read from %s: %s\n", peer, err)
		}
		
		go handlePeer(peer)	
	}
}

func handlePeer(p *Peer) {
	p.mux.Lock()
	defer p.mux.Unlock()

	typ, err := p.pakt.ReadByte()

	if err != nil {
		return
	}

	switch shared.MessageType(typ) {
	case shared.GetPeersMessage:
		res := server.PeersString()
		p.pakt.Flush()
		p.pakt.WriteString(res)
		server.Send(p.pakt.GetBytes(), p.String())
	case shared.PingMessage:
		p.ping <- struct{}{} 
	default:
		// TODO: Handle Invalid Message Type
	}
}
