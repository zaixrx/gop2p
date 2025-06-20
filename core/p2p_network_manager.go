package p2p

import (
	tr "github.com/zaixrx/gop2p/core/transport"
)

type network_manager struct {
	transport tr.Transport
}

func NewNetworkManager() *network_manager {
	return &network_manager{
		transport: &tr.TCPTransport{}, // TODO: Implement UDP transport
	}
}

func (nm *network_manager) Listen(addr string) error {
	return nm.transport.Listen(addr)
}

func (nm *network_manager) Accept() (tr.Conn, error) {
	conn, err := nm.transport.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (nm *network_manager) Connect(addr string) (tr.Conn, error) {
	return nm.transport.Connect(addr)
}

func (nm *network_manager) Close() error {
	return nm.transport.Close()
}
