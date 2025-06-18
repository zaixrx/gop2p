package shared

import (
	"encoding/binary"
	"strings"
	"fmt"
)

type Packet struct {
	consume bool
	data []byte
	offset uint32  
}

func NewPacket() *Packet {
	return &Packet{
		consume: true,
		data: make([]byte, 0),
		offset: 0,
	}
}
func (p *Packet) SetConsume(consume bool) {
	p.consume = consume
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
	if p.offset + n > length { 
		return nil, fmt.Errorf("ERROR: out of bound by %d elements", p.offset + n - length)
	}
	dat := p.data[p.offset : n + p.offset]
	if p.consume {
		p.offset += n
	}
	return dat, nil
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
func (p *Packet) ReadStringArr() ([]string, error) {
	arrRaw, err := p.ReadString()
	if err != nil {
		return nil, nil
	}
	return strings.Split(arrRaw, " "), nil
}
func (p *Packet) ReadPool() (*PublicPool, error) {
	poolID, err := p.ReadString()
	if err != nil {
		return nil, err
	}
	hostIP, err := p.ReadString()
	if err != nil {
		return nil, err
	}
	yourIP, err := p.ReadString()
	if err != nil {
		return nil, err
	}
	peerIPs, err := p.ReadStringArr()
	if err != nil {
		return nil, err
	}
	return &PublicPool{
		Id: poolID,
		HostIP: hostIP,
		YourIP: yourIP,
		PeerIPs: peerIPs,
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
func (p *Packet) WriteStringArr(dat []string) error {
	return p.WriteString(strings.Join(dat, " "))
}
func (p *Packet) WritePool(dat *PublicPool) error {
	p.WriteString(dat.Id)
	p.WriteString(dat.HostIP)
	p.WriteString(dat.YourIP)
	p.WriteStringArr(dat.PeerIPs)
	return nil
}
func (p *Packet) WriteBytesBefore(dat []byte) error {
	p.data = append(dat, p.data...)
	return nil
}
func (p *Packet) WriteBytesAfter(dat []byte) error {
	p.data = append(p.data, dat...)
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
