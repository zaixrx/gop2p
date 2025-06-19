package P2P

import (
	"log"
)

type network_manager struct {
	listening bool
	transport t_transport
}

func NewNetworkManager() *network_manager {
	return &network_manager{
		transport: &tcp_transport{}, // TODO: Implement UDP transport
	}
}

func (nm *network_manager) Listen(addr string) error {
	nm.listening = true 
	log.Println("Called Listen")
	return nm.transport.Listen(addr)
}

func (nm *network_manager) Accept() (t_conn, error) {
	log.Println("Called Accept")
	conn, err := nm.transport.Accept()
	if err != nil {
		return nil, err
	}
	log.Println("Accepted Conn")
	return conn, nil
}

func (nm *network_manager) Connect(addr string) (t_conn, error) {
	log.Println("Called Connect")
	return nm.transport.Connect(addr)
}

func (nm *network_manager) Close() error {
	nm.listening = false
	log.Println("Called Close")
	return nm.transport.Close()
}
