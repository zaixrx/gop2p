package main

import (
	"context"
)

type StateAction[State any] func(context.Context, *State) (StateAction[State], error)

type StateContext[State any] struct {
 	initialAction StateAction[State]
	backgroundTask StateAction[State]
	state State
}

func NewStateContext(initialAction, backgroundTask StateAction[State]) *StateContext[State] {
	return &StateContext[State]{
		initialAction: initialAction,
		backgroundTask: backgroundTask,
		state: State{},
	}
}

func (sc *StateContext[State]) Run() (*State, error) {
	var (
		currSA StateAction[State] = sc.initialAction
		firstTime = false
		err error
	)

	ctx, cancel:= context.WithCancel(context.Background())
	defer cancel()

	for {
		currSA, err = currSA(ctx, &sc.state)
		if err != nil {
			return nil, err
		}
		if currSA == nil {
			break
		}
		// Looks ugly but works :\
		if !firstTime {
			go sc.backgroundTask(ctx, &sc.state)
			firstTime = true
		}
		// d := sc.decisionReader(Keys(saMap))
		// currSA = saMap[d]
	}
	return &sc.state, nil
}

// func Keys[T any](m map[string]T) []string {
// 	keys := make([]string, len(m))
// 	i := 0
// 	for key := range m {
// 		keys[i] = key
// 		i++
// 	}
// 	return keys
// }
