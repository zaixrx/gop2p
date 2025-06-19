package main
import (
	"github.com/zaixrx/gop2p/broadcaster"
)

var server *broadcaster.Server

const (
	Port int = 6969
	PoolPingTimeout int = 5
	Hostname string = "127.0.0.1"
)

func main() {
	server = broadcaster.NewServer(PoolPingTimeout)
	server.Start(Hostname, Port)
	for {
		err := server.Listen()
		if err != nil {
			continue
		}
	}
}
