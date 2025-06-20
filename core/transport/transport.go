package transport

import (
	"github.com/zaixrx/gop2p/shared"
)

type Packet = shared.Packet

func NewPacket() *Packet {
	return shared.NewPacket()
}

type Transport interface {
	Listen(string) error
	Accept() (Conn, error)
	Connect(string) (Conn, error)
	Close() error
}

type Conn interface {
	Write(*Packet) (int, error)
	Read() (*Packet, error)
	Address() string
	Close() error
}
