package transport

import (
	"fmt"
	"net"
)

type TCPTransport struct {
	listener net.Listener
}

type TCPConn struct {
	conn net.Conn
	buff []byte
}

func (tcp *TCPTransport) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tcp.listener = listener
	return nil
}

func (tcp *TCPTransport) Accept() (Conn, error) {
	conn, err := tcp.listener.Accept()
	if err != nil {
		return nil, err
	}
	return newTcpConn(conn), nil
}

func (tcp *TCPTransport) Connect(addr string) (Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return newTcpConn(conn), nil
}

func (tcp *TCPTransport) Close() error {
	return tcp.listener.Close()
}

func newTcpConn(conn net.Conn) *TCPConn {
	return &TCPConn{
		conn: conn,
		buff: make([]byte, 1024),
	}
}

func (conn *TCPConn) Write(packet *Packet) (int, error) {
	return conn.conn.Write(packet.GetBytes())
}

func (conn *TCPConn) Read() (*Packet, error) {
	nbr, err := conn.conn.Read(conn.buff)
	if nbr == 0 || err != nil {
		fmt.Println(nbr, len(conn.buff))
		return nil, fmt.Errorf("tcp socket sent fin flag") 
	}

	packet := NewPacket()
	packet.Load(conn.buff[:nbr])

	return packet, nil
}

func (conn *TCPConn) Address() string {
	return conn.conn.RemoteAddr().String()
}

func (conn *TCPConn) Close() error {
	return conn.conn.Close()
}
