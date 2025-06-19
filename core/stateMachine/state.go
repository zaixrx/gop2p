package stateMachine 

import (
	"log"
	"context"
)

// I created a state machine system because I needed to listen to certain events
// on a fixed chronological order, I know it poses certain limitations but atleast structure of programming
// to the use but hey! that's how the tool is made(And I'm too far into this shit to just scrap the idea)
type StateJob[t_job_state any] func(ctx context.Context, s *t_job_state) (StateJob[t_job_state], error)

type StateMachine[t_job_state any] struct {
 	initJob StateJob[t_job_state]
	state t_job_state 
}

func NewStateMachine[t_job_state any](initState t_job_state, initJob StateJob[t_job_state]) *StateMachine[t_job_state] {
	return &StateMachine[t_job_state]{
		state: initState,
		initJob: initJob,
	}
}

func (sm *StateMachine[t_job_state]) GetState() t_job_state {
	return sm.state
}

// This must be run in the main goroutine as the return type imposes
func (sm *StateMachine[t_job_state]) Run(ctx context.Context) {
	var (
		job StateJob[t_job_state] = sm.initJob
		err error
	)
	
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job, err = job(ctx, &sm.state)
			if err != nil {
				log.Println(err)
				// TODO: make onError event subscriber
			}
			if job == nil {
				return 
			}
			i++
		}
	}	
}
