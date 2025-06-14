package main

import (
	"net"
	"log"
	"p2p/shared"
)

type MessageHandler func(*shared.Packet, *Broadcast) error 

type Broadcast struct {
	conn net.Conn
}

func NewBroadcast(conn net.Conn) *Broadcast {
	return &Broadcast{
		conn: conn,
	}
}

func (b *Broadcast) Listen() (*shared.Packet, error) {
	buff := make([]byte, 1024)
	n, err := b.conn.Read(buff)
	if err != nil {
		return nil, err
	}
	packet := shared.NewPacket()
	packet.Load(buff[:n])
	return packet, nil 
}

func (b *Broadcast) Write(dat []byte) (int, error) {
	return b.conn.Write(dat)
}

func (b *Broadcast) SendPoolRetreivalMessage() error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageRetrievePools))
	_, err := b.Write(packet.GetBytes())
	return err
}

func (b *Broadcast) SendPoolCreateMessage() error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageCreatePool))
	_, err := b.Write(packet.GetBytes())
	return err
}

func (b *Broadcast) SendPoolJoinMessage(poolID string) error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageJoinPool))
	packet.WriteString(poolID)
	_, err := b.Write(packet.GetBytes())
	return err
}

func (b *Broadcast) SendPoolLeaveMessage(poolID string) error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessageLeavePool))
	packet.WriteString(poolID)
	_, err := b.Write(packet.GetBytes())
	return err
}

func (b *Broadcast) SendPoolPingMessage(poolID string) error {
	packet := shared.NewPacket()
	packet.WriteByte(byte(shared.MessagePoolPing))
	packet.WriteString(poolID)
	_, err := b.Write(packet.GetBytes())
	log.Println("Sending Ping Message")
	return err
}
