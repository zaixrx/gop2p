package P2P 

import (
	"p2p/shared"
)

type P2PPacket = shared.Packet

func NewPacket() *P2PPacket {
	return shared.NewPacket()
}

type P2PMessage struct {
	packet *P2PPacket
}

type t_transport interface {
	Listen(string) error
	Accept() (t_conn, error)
	Connect(string) (t_conn, error)
	Close()
}

type t_conn interface {
	Write(*P2PMessage) (int, error)
	Read() (*P2PMessage, error)
	Address() string
}
