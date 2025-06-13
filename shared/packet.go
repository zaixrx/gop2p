package shared

import (
	"encoding/binary"
	"strings"
	"fmt"
)

type Packet struct {
	data []byte
	offset uint32  
}

func NewPacket() *Packet {
	return &Packet{
		data: make([]byte, 0),
		offset: 0,
	}
}
func (p *Packet) Load(data []byte) {
	p.data = make([]byte, len(data))
	copy(p.data, data)
	p.offset = 0
}
func (p *Packet) Get(n uint32) ([]byte, error) {
	length := uint32(len(p.data))
	if length == 0 {
		return nil, fmt.Errorf("ERROR: attempting to read empty packet")
	}
	p.offset += n
	if p.offset > length { 
		p.offset = length - 1
		return nil, fmt.Errorf("ERROR: out of bound by %d\n", p.offset - length)
	}
	return p.data[p.offset - n:p.offset], nil
}
func (p *Packet) ReadByte() (byte, error) {
	dat, err := p.Get(1) 
	if err != nil {
		return 0, err
	}
	return dat[0], nil
}
func (p *Packet) ReadUInt32() (uint32, error) {
	dat, err := p.Get(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(dat), nil
}
func (p *Packet) ReadString() (string, error) {
	n, err := p.ReadUInt32()
	if err != nil {
		return "", err
	}
	dat, err := p.Get(n)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}
func (p *Packet) ReadPool() (*Pool, error) {
	id, err := p.ReadString()
	if err != nil {
		return nil, err
	}
	host, err := p.ReadString()
	if err != nil {
		return nil, err
	}
	peers, err := p.ReadString()
	if err != nil {
		return nil, err
	}
	
	return &Pool{
		id: id,
		host: host,
		peers: strings.Split(peers, " "),
	}, nil
}

// TODO: Make read and write mode(packet is either for read or for writing)
func (p *Packet) WriteByte(dat byte) error {
	p.data = append(p.data, dat)
	return nil
}
func (p *Packet) WriteUint32(dat uint32) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, dat)
	p.data = append(p.data, buf...)
	return nil
}
func (p *Packet) WriteString(dat string) error {
	p.WriteUint32(uint32(len(dat)))
	p.data = append(p.data, []byte(dat)...)
	return nil
}
func (p *Packet) WritePool(key string, pool *Pool) error {
	p.WriteString(key)
	p.WriteString(pool.host)
	p.WriteString(strings.Join(pool.peers, " "))
	return nil
}
func (p *Packet) GetBytes() []byte {
	dat := p.data
	p.Flush()
	return dat 
}
func (p *Packet) Flush() {
	p.data = []byte{} 
	p.offset = 0
}
