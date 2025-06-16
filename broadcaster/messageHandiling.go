package main

import (
	"net"
	"fmt"
	"p2p/shared"
	"github.com/google/uuid"
)

func HandlePoolCreateMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	key := uuid.New().String()
	pool := shared.NewPool(key, addr)
	server.pools[key] = pool 
	go server.MonitorPool(key)

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageJoinPool))
	p.WritePool(pool.ToPublic(addr))

	server.Write(p.GetBytes(), addr)
	
	return nil
}

func HandlePoolRetreivalMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	p := shared.NewPacket()
	p.WriteByte(byte(shared.MessageRetrievePools))
	keys := make([]string, len(server.pools))
	i := 0
	for key := range server.pools {
		keys[i] = key
		i++
	}

	fmt.Println("Write String Arr")

	err := p.WriteStringArr(keys)
	keys, err = p.ReadStringArr()
	err = p.WriteStringArr(keys)
	
	fmt.Println("Wrote String Arr")

	if err != nil {
		return err
	}
	
	server.Write(p.GetBytes(), addr)

	return nil
}

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
	p.WritePool(pool.ToPublic(addr))

	server.Write(p.GetBytes(), addr)
	
	return nil
}

func HandlePoolPingMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
	server.mux.Lock()
	defer server.mux.Unlock()

	poolID, err := packet.ReadString()
	if err != nil {
		return err
	}

	pool, err := server.GetPool(poolID)
	if err != nil {
		return err 
	}

	if addr.String() != pool.Peers[pool.HostID].String() {
		return fmt.Errorf("ERROR: only the host can send ping messages")
	}

	pool.PingChan<-struct{}{}

	return nil
}
