package P2P

import (
	"net"
	"fmt"
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
	return newTcpConn(conn), nil
}

func (tcp *tcp_transport) Connect(addr string) (t_conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return newTcpConn(conn), nil
}

func (tcp *tcp_transport) Close() error {
	return tcp.listener.Close()
}

func newTcpConn(conn net.Conn) *tcp_conn {
	return &tcp_conn{
		conn: conn,
		buff: make([]byte, 1024),
	}
}

func (conn *tcp_conn) Write(packet *Packet) (int, error) {
	fmt.Println("Connection Write")
	return conn.conn.Write(packet.GetBytes())
}

func (conn *tcp_conn) Read() (*Packet, error) {
	nbr, err := conn.conn.Read(conn.buff)
	if nbr == 0 || err != nil {
		fmt.Println(nbr, len(conn.buff))
		return nil, fmt.Errorf("tcp socket sent fin flag!") 
	}

	packet := NewPacket()
	packet.Load(conn.buff[:nbr])

	fmt.Println("Connection Read")

	return packet, nil
}

func (conn *tcp_conn) Address() string {
	fmt.Println("Connection Address")
	return conn.conn.RemoteAddr().String()
}

func (conn *tcp_conn) Close() error {
	fmt.Println("Connection Closed")
	return conn.Close()
}
