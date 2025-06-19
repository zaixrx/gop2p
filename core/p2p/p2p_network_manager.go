package p2p

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
	return nm.transport.Listen(addr)
}

func (nm *network_manager) Accept() (t_conn, error) {
	conn, err := nm.transport.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (nm *network_manager) Connect(addr string) (t_conn, error) {
	return nm.transport.Connect(addr)
}

func (nm *network_manager) Close() error {
	nm.listening = false
	return nm.transport.Close()
}
