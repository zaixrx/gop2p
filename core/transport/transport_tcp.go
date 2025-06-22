package transport

import (
	"sync"
	"fmt"
	"net"
	"encoding/binary"
)

type TCPTransport struct {
	listening bool
	listener net.Listener
}

type TCPConn struct {
	// read & writes must be atomic
	wm, rm sync.Mutex
	conn net.Conn
	buff []byte
}

func (tcp *TCPTransport) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tcp.listener = listener
	tcp.listening = true
	return nil
}

func (tcp *TCPTransport) Accept() (Conn, error) {
	if !tcp.listening {
		return nil, fmt.Errorf("tcp_transport is not listening")
	}
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
	if !tcp.listening {
		return fmt.Errorf("tcp_transport is not listening")
	}
	tcp.listening = false
	return tcp.listener.Close()
}

func newTcpConn(conn net.Conn) *TCPConn {
	return &TCPConn{
		conn: conn,
		buff: make([]byte, 1024),
	}
}

func (conn *TCPConn) write_full(dat []byte) (int, error) {
	var (
		nbw = 0
	)

	for nbw < len(dat) {
		n, err := conn.conn.Write(dat[nbw:])
		nbw += n

		if err != nil {
			return nbw, err
		}
	}

	return nbw, nil
}

func (conn *TCPConn) Write(packet *Packet) (int, error) {
	dat := packet.GetBytes()

	prelenbuff := make([]byte, 4)
	binary.LittleEndian.PutUint32(prelenbuff, uint32(len(dat)))

	conn.wm.Lock()
	defer conn.wm.Unlock()

	// write prelength
	n, err := conn.write_full(prelenbuff);
	if err != nil {
		return n, err
	}

	// write the main data 
	n1, err := conn.write_full(dat);
	return n + n1, err 
}

func (conn *TCPConn) read_n(n int) ([]byte, error) {
	var (
		out = make([]byte, n)
	)

	for n > 0 {
		i, err := conn.conn.Read(out)
		n -= i

		if err != nil {
			return out, err
		}
	}

	return out, nil
}

func (conn *TCPConn) Read() (*Packet, error) {
	conn.rm.Lock()
	defer conn.rm.Unlock()

	prelenbuff, err := conn.read_n(4)
	if err != nil {
		return nil, err
	}
	dat, err := conn.read_n(int(binary.LittleEndian.Uint32(prelenbuff)))
	if err != nil {
		return nil, err
	}
	packet := NewPacket()
	packet.Load(dat)
	return packet, nil
}

func (conn *TCPConn) Address() string {
	return conn.conn.RemoteAddr().String()
}

func (conn *TCPConn) Close() error {
	return conn.conn.Close()
}
