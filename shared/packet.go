package shared

import (
	"encoding/binary"
	"fmt"
	"slices"
)

type Packet struct {
	data []byte
	sender string
	offset uint32  
}

func NewPacket(sender string) *Packet {
	return &Packet{
		sender: sender,
	}
}
func (p *Packet) Load(data []byte) {
	p.data = make([]byte, len(data))
	copy(p.data, data)
	fmt.Println("Comparing Slices", slices.Compare(p.data, data), "Length", len(p.data))
	p.offset = 0
}
func (p *Packet) Get(n uint32) ([]byte, error) {
	p.offset += n
	length := uint32(len(p.data))
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
	fmt.Println("Managed to read uint32")
	return binary.LittleEndian.Uint32(dat), nil
}
func (p *Packet) ReadString() (string, error) {
	n, err := p.ReadUInt32()
	if err != nil {
		return "", err
	}
	fmt.Println("String length is", n)
	dat, err := p.Get(n)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}


// TODO: Make read and write mode(packet is either for read or for writing)
func (p *Packet) WriteByte(dat byte) {
	p.data = append(p.data, dat)
}
func (p *Packet) WriteUint32(dat uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, dat)
	p.data = append(p.data, buf...)
}
func (p *Packet) WriteString(dat string) {
	p.WriteUint32(uint32(len(dat)))
	p.data = append(p.data, []byte(dat)...)
}
func (p *Packet) GetBytes() []byte {
	return p.data
}
func (p *Packet) Flush() {
	p.data = p.data[:0]
}
