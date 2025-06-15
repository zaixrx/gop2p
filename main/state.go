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

func (sc *StateContext[State]) Run(cancelChan chan<- context.CancelFunc) (*State, error) {
	var (
		currSA StateAction[State] = sc.initialAction
		err error
	)

	ctx, cancel := context.WithCancel(context.Background())	

	if cancelChan != nil {
		cancelChan <- cancel
	}
	defer cancel() 
	
	i := 0
	for {
		select {
		case <-ctx.Done():
			return &sc.state, ctx.Err()
		default:
			currSA, err = currSA(ctx, &sc.state)
			if err != nil {
				log.Printf("ERROR %d: %s", i, err)
				// TODO: make onError event subscriber
			}
			if currSA == nil {
				return &sc.state, err
			}
			i++
		}
	}	
}
