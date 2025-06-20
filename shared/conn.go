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

func (mt MessageType) String() string {
	return [...]string{"CreatePool", "Pool Ping", "Retrieve Pools", "Join Pool", "Error", "Pool Ping Timeout"}[mt]
}
