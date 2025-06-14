package shared

const (
	Port = 6969
	Hostname = "127.0.0.1"
	PoolPingTicks = 5 
	PoolPingTimeout = 1 
)

type MessageType byte

const (
	// Client to Server
	MessageCreatePool MessageType = iota
	MessagePoolPing
	MessageLeavePool
	
	// Hybrid Messages
	MessageRetrievePools
	MessageJoinPool

	// Server to Client
	MessagePoolPingTimeout
)

