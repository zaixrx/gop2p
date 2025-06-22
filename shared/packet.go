package shared 

import (
	"encoding/binary"
	"strings"
	"fmt"
)

type Packet struct {
	data []byte
	offset uint32  
	writeBefore bool
}

const str_sep string = " " 
func joinStrArr(arr []string) string {
	return strings.Join(arr, str_sep)
}
func splitStr(str string) []string {
	if len(str) == 0 {
		return []string{}
	}
	return strings.Split(str, str_sep)
}

func NewPacket() *Packet {
	return &Packet{
		writeBefore: false,
		data: make([]byte, 0),
		offset: 0,
	}
}

func (p *Packet) GetLen() int {
	return len(p.data)
}
func (p *Packet) SetWriteBefore(val bool) {
	p.writeBefore = val
}

func (p *Packet) Load(data []byte) {
	p.data = make([]byte, len(data))
	copy(p.data, data)
	p.offset = 0
}
func (p *Packet) GetOffset() uint32 {
	return p.offset
}
func (p *Packet) SetOffset(val uint32) {
	p.offset = val
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
	p.offset += n
	return dat, nil
}
func (p *Packet) ReadByte() (byte, error) {
	dat, err := p.Get(1) 
	if err != nil {
		return 0, err
	}
	return dat[0], nil
}
func (p *Packet) ReadUInt16() (uint16, error) {
	dat, err := p.Get(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(dat), nil 
}
func (p *Packet) ReadUInt32() (uint32, error) {
	dat, err := p.Get(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(dat), nil
}
func (p *Packet) ReadUInt64() (uint64, error) {
	dat, err := p.Get(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(dat), nil
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
	return splitStr(arrRaw), nil
}
func (p *Packet) ReadPool() (*PublicPool, error) {
	arr, err := p.ReadStringArr()
	if err != nil {
		return nil, err
	}
	if len(arr) < 4 {
		return nil, fmt.Errorf("cannot decode invalid packet(ReadPool)")
	}

	return &PublicPool{
		Id: arr[0],
		HostIP: arr[1],
		YourIP: arr[2],
		PeerIPs: arr[3:],
	}, nil
}

// TODO: Make read and write mode(packet is either for read or for writing)
func (p *Packet) appnd(dat []byte) error {
	if p.writeBefore {
		p.data = append(dat, p.data...)
	} else {
		p.data = append(p.data, dat...)
	}
	return nil
}

func (p *Packet) WriteByte(dat byte) error {
	p.appnd([]byte{dat})
	return nil
}
func (p *Packet) WriteBytes(dat []byte) error {
	p.appnd(dat)
	return nil
}

func getUint16(dat uint16) []byte {
	buf := make([]byte, 4) // TODO: this is stupid bad
	binary.LittleEndian.PutUint16(buf, dat)
	return buf
}
func (p *Packet) WriteUint16(dat uint16) error {
	return p.appnd(getUint16(dat))
}

func getUint32(dat uint32) []byte {
	buf := make([]byte, 4) // TODO: this is stupid bad
	binary.LittleEndian.PutUint32(buf, dat)
	return buf
}
func (p *Packet) WriteUint32(dat uint32) error {
	return p.appnd(getUint32(dat))
}

func getUint64(dat uint64) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint64(buf, dat)
	return buf
}
func (p *Packet) WriteUint64(dat uint64) error {
	return p.appnd(getUint64(dat))
}

func getString(dat string) []byte {
	return append(getUint32(uint32(len(dat))), []byte(dat)...)
}
func (p *Packet) WriteString(dat string) error {
	return p.appnd(getString(dat))
}
func (p *Packet) WriteStringArr(dat []string) error {
	return p.appnd(getString(joinStrArr(dat)))
}

func (p *Packet) WritePool(dat *PublicPool) error {
	return p.WriteStringArr([]string{dat.Id, dat.HostIP, dat.YourIP, joinStrArr(dat.PeerIPs)})
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
