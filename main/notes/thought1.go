/*

What I want to make: 

1) Make it easy and intuitive to listen to external events
2) Be able to spit in and out values with ease in an organized manner
3) Terminating a State Machine:
	- Terminate on a TerminateError
	- Be able to manually terminate a state machine using context 
	(Note that using context makes it possible to have a group of goroutines that terminate with
	each other)
*/

package main

func main() {
	br := BroadcastManager(takeRoute)

	/*
	In an action that requires users command
	you wait for the user to provide the action by providing
	
	type ExternalAction struct {
		msgType MessageType
		args []string // Only strings for now
	}

	func (s *State) ListenExternal(toWhat []MessageType) ExternalAction {
		for {
			externAct<-externChan
			if slices.Contains(toWhat, externAct.msgType) {
				return externAct 
			}
		}
	}

	What I want to make: 

	1) Make it easy and intuitive to listen to external events
	2) By able to spit in and out values with ease in an organized manner
	3) Terminating a State Machine:
		- Or terminate on a TerminateError
		- Be able to manually terminate a state machine using context 
		(Note that using context makes it possible to have a group of goroutines that terminate with
		each other)
	
	func HandleRetreivePool(ctx context.Context, state *BRState) (StateAction[BRState], error) {
		rawPools, err := state.packet.ReadString()
		if err != nil {
			return nil, err
		}

		state.Pools = strings.Split(rawPools, " ") 
		
		action := ListenExternal([]MessageType{MessageType.JoinPool, MessageType.CreatePool})
		switch action.type {
		case MessageType.JoinPool:
			state.br.SendJoinPool(action.args[0])
		case MessageType.CreatePool:
			state.br.SendCreatePool()
		}

		Listen([]MessageType{MessageType.HandleJoinPool})
	}

	func JoinPool(poolID string) {
			
	}

	that function will wait for a channel userEv
	*/

	brResult := br.Run(stateChan)

	pool := brResult.Pool
	brResult.Ping()
}

func takeRoute() (typ MessageType, args []string)
