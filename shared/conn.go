package shared

const (
	Port = 6969
	Hostname = "127.0.0.1"
	PingTicks = 10 
	MaxPingTimeout = 1
)

type MessageType byte

const (
	GetPeersMessage MessageType = iota
	PingMessage
)

