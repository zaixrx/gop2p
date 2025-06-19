package p2p

import (
	"github.com/zaixrx/gop2p/shared"
)

type Packet = shared.Packet

func NewPacket() *Packet {
	return shared.NewPacket()
}

type t_transport interface {
	Listen(string) error
	Accept() (t_conn, error)
	Connect(string) (t_conn, error)
	Close() error
}

type t_conn interface {
	Write(*Packet) (int, error)
	Read() (*Packet, error)
	Address() string
	Close() error
}
