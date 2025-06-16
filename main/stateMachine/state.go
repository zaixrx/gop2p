package stateMachine 

import (
	"log"
	"context"
)

// I created a state machine system because I needed to listen to certain events
// on a fixed chronological order, I know it poses certain limitations but atleast structure of programming
// to the use but hey! that's how the tool is made(And I'm too far into this shit to just scrap the idea)
type StateJob[t_state, t_output any] func(s *t_state, o *t_output) (StateJob[t_state, t_output], error)

type StateMachine[t_state, t_output any] struct {
 	initJob StateJob[t_state, t_output]
	output t_output 
	state t_state 
}

func NewStateMachine[state, output any](initState state, initJob StateJob[state, output]) *StateMachine[state, output] {
	return &StateMachine[state, output]{
		state: initState,
		initJob: initJob,
	}
}

func (sm *StateMachine[t_state, t_output]) GetState() t_state {
	return sm.state
}

func (sm *StateMachine[t_state, t_output]) GetOutput() t_output {
	return sm.output
}

// This must be run in the main goroutine as the return type imposes
func (sm *StateMachine[t_state, t_output]) Run() {
	var (
		job StateJob[t_state, t_output] = sm.initJob
		err error
	)

	ctx, cancel := context.WithCancel(context.Background())	
	defer cancel() 
	
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job, err = job(&sm.state, &sm.output)
			if err != nil {
				log.Printf("ERROR %d: %s", i, err)
				// TODO: make onError event subscriber
			}
			if job == nil {
				return 
			}
			i++
		}
	}	
}
