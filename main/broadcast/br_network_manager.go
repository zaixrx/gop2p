package Broadcast 

import (
	"net"
	"strconv"
	"p2p/shared"
)

type NetworkManager struct {
	port uint16
	hostName string
	conn net.Conn
}

func NewNetworkManager(hostName string, port uint16) *NetworkManager {	
	return &NetworkManager{
		hostName: hostName,
		port: port,
	}
}

func (nw *NetworkManager) Connect() error {
	addr := net.JoinHostPort(shared.Hostname, strconv.Itoa(shared.Port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	nw.conn = conn
	return nil
}

func (nw *NetworkManager) Close() error {
	return nw.conn.Close()
}

func (nw *NetworkManager) Listen() (*shared.Packet, error) {
	buff := make([]byte, 1024)
	n, err := nw.conn.Read(buff)
	if err != nil {
		return nil, err
	}
	packet := shared.NewPacket()
	packet.Load(buff[:n])
	return packet, nil 
}

func (nm *NetworkManager) Write(packet *shared.Packet) (int, error) {
	byt := packet.GetBytes()
	return nm.conn.Write(byt)
}

func (nm *NetworkManager) SendRetrievePools() error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageRetrievePools))
	_, err := nm.Write(packet)
	return err
}

func (nm *NetworkManager) SendCreatePool() error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageCreatePool))
	_, err := nm.Write(packet)
	return err
}

func (nm *NetworkManager) SendJoinPool(poolID string) error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageJoinPool))
	packet.WriteString(poolID)
	_, err := nm.Write(packet)
	return err
}

func (nm *NetworkManager) SendPoolPingMessage(poolID string) error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessagePoolPing))
	packet.WriteString(poolID)
	_, err := nm.Write(packet)
	return err
}
