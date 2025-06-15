package shared

const (
	Port = 6969
	Hostname = "127.0.0.1"
	PoolPingTicks = 5 
	PoolPingTimeout = 1 
	MaxPCR = 10 // Pool Connection Requests
)

type MessageType byte

const (
	// Client to Server
	MessageCreatePool MessageType = iota
	MessagePoolPing
	
	// Hybrid Messages
	MessageRetrievePools
	MessageJoinPool

	// Server to Client
	MessageError	
	MessagePoolPingTimeout
)

