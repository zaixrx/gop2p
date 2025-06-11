package shared

const (
	Port = 6969
	Hostname = "127.0.0.1"
)

type MessageType byte

const (
	GetPeers MessageType = iota
)

