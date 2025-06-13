package main

import (
	"fmt"
	"net"
	"p2p/shared"
	"strings"
	"strconv"
	//"github.com/google/uuid"
)

func HandlePoolJoinMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	key, err := packet.ReadString() 
	if err != nil {
		return err
	}
	pool, err := server.GetPool(key)	
	if err != nil {
		return err
	}
	pool.Add(addr)

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageJoinPool))
	p.WriteString(key)
	p.WriteString(pool.host)
	keys := make([]string, len(pool.peers))
	i := 0
	for key := range pool.peers {
		keys[i] = key
		i++
	}
	p.WriteString(strings.Join(keys, " "))

	server.Write(p.GetBytes(), addr)
	
	return nil
}

func HandlePoolLeaveMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	key, err := packet.ReadString() 
	if err != nil {
		return err
	}
	pool, err := server.GetPool(key)	
	if err != nil {
		return err
	}
	err = pool.Remove(addr)
	if err != nil {
		return err
	}

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageJoinPool))
	p.WriteString(key)

	server.Write(p.GetBytes(), addr)
	
	return nil
}

func HandlePoolCreateMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	pool := NewPool(addr)
	//key := uuid.New().String()
	key := strconv.Itoa(len(server.pools) + 1)
	server.pools[key] = pool

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageCreatePool))
	p.WriteString(key)

	server.Write(p.GetBytes(), addr)
	
	return nil
}

func HandlePoolDeleteMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	key, err := packet.ReadString()
	if err != nil {
		return err
	}
	_, ok := server.pools[key]
	if !ok {
		return fmt.Errorf("ERROR: cannot delete inexisting pool")
	}
	delete(server.pools, key)

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageDeletePool))
	p.WriteString(key)

	server.Write(p.GetBytes(), addr)

	return nil
}

func HandlePoolRetreivalMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageRetreivePools))
	keys := make([]string, len(server.pools))
	i := 0
	for key := range server.pools {
		keys[i] = key
		i++
	}
	p.WriteString(strings.Join(keys, " "))
	b := p.GetBytes()
	nbw, err := server.Write(b, addr)
	if err != nil || nbw < len(b) {
		return fmt.Errorf("ERROR: message failed to be sent")
	}
	return nil
}
