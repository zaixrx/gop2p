package P2P

type network_manager struct {
	transport t_transport
}

func NewNetworkManager() *network_manager {
	return &network_manager{
		transport: &tcp_transport{}, // TODO: Implement UDP transport
	}
}

func (nm *network_manager) Listen(addr string) error {
	return nm.transport.Listen(addr)
}

func (nm *network_manager) Accept() (t_conn, error) {
	return nm.transport.Accept()
}

func (nm *network_manager) Connect(addr string) (t_conn, error) {
	return nm.transport.Connect(addr)
}

func (nm *network_manager) Close() {
	nm.transport.Close()
}
