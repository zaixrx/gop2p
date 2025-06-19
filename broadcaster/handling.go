package broadcaster

import (
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/zaixrx/gop2p/shared"
)

func handlePoolCreateMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
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

func handlePoolRetreivalMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
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

	err := p.WriteStringArr(keys)

	if err != nil {
		return err
	}
	
	server.Write(p.GetBytes(), addr)

	return nil
}

func handlePoolJoinMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
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

func handlePoolPingMessage(packet *shared.Packet, addr *net.UDPAddr, server *Server) error {
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
