package P2P

import (
	"net"
)

type tcp_transport struct {
	listener net.Listener
}

type tcp_conn struct {
	conn net.Conn
	buff []byte
}

func (tcp *tcp_transport) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tcp.listener = listener
	return nil
}

func (tcp *tcp_transport) Accept() (t_conn, error) {
	conn, err := tcp.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &tcp_conn{
		conn: conn,
	}, nil
}

func (tcp *tcp_transport) Connect(addr string) (t_conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &tcp_conn{
		conn: conn,
	}, nil
}

func (tcp *tcp_transport) Close() {
	tcp.listener.Close()
}

func (conn *tcp_conn) Write(msg *P2PMessage) (int, error) {
	return conn.conn.Write(msg.packet.GetBytes())
}

func (conn *tcp_conn) Read() (*P2PMessage, error) {
	nbr, err := conn.conn.Read(conn.buff)
	if err != nil {
		return nil, err
	}

	packet := NewPacket()
	packet.Load(conn.buff[:nbr])

	return &P2PMessage{
		packet: packet,
	}, nil
}

func (conn *tcp_conn) Address() string {
	return conn.conn.RemoteAddr().String()
}
