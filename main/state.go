package main

import (
	"log"
	"context"
)

// I created a state machine system because I needed to listen to certain events
// on a fixed chronological order, I know it poses certain limitations but atleast structure of programming
// to the use but hey! that's how the tool is made(And I'm too far into this shit to just scrap the idea)
type StateAction[State any] func(context.Context, *State) (StateAction[State], error)

type StateContext[State any] struct {
 	initialAction StateAction[State]
	state State
}

func NewStateContext[State any](initialAction StateAction[State]) *StateContext[State] {
	return &StateContext[State]{
		initialAction: initialAction,
	}
}

func (sc *StateContext[State]) Run(stateChan chan<- *State, cancelChan chan<- context.CancelFunc) { 
	var (
		currSA StateAction[State] = sc.initialAction
		err error
	)

	ctx, cancel := context.WithCancel(context.Background())	

	if cancelChan != nil {
		cancelChan <- cancel
	}

	defer func() {
		if stateChan != nil {
			stateChan <- &sc.state	
		}
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			currSA, err = currSA(ctx, &sc.state)
			if err != nil {
				log.Printf("ERROR: %s", err)
				return	
			}
			if currSA == nil {
				return
			}
		}
	}	
}
