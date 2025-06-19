package shared 

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

