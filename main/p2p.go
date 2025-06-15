package main

import (
	"net"
	"fmt"
	"log"
	"context"
	"p2p/shared"
)

//type StateAction[State any] func(context.Context, *State) (StateAction[State], error)

type P2PState struct {
	Pool *shared.PublicPool
	Conns map[string]*net.Listener
}

func CreateP2PStage(pool *shared.PublicPool) *StateContext[P2PState] {
	sc := NewStateContext(ConnectToPeers)
	sc.state.Pool = pool
	return sc
}

func ConnectToPeers(ctx context.Context, state *P2PState) (StateAction[P2PState], error) {
	for pi := range state.Pool.PeerIPs {
		var (
			peerIP string = state.Pool.PeerIPs[pi]
			list net.Listener
			err error
			i = 0
		)

		if peerIP == state.Pool.YourIP {
			continue
		}

		for i < shared.MaxPCR {
			list, err = net.Listen("tcp", peerIP)
			if err != nil {
				i++
				continue
			}
			log.Printf("Failed to connect to %s\n", peerIP)
			break
		}
		
		if i < shared.MaxPCR {
			state.Conns[peerIP] = &list
		}
	}

	if len(state.Conns) < len(state.Pool.PeerIPs) {
		return nil, fmt.Errorf("Failed to connect to any peer\n", )
	}

	return nil, nil
}
