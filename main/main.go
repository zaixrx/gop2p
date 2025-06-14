package main

import (
	"os"
	"log"
	"bufio"
	"context"
)

func main() {
	// That is really how simple it is
	// I'm a genuis man!!
	stage := CreateBroadcastingStage(BackgroundTask)
	state, err := stage.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(state.currentPool)
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func BackgroundTask(ctx context.Context, state *State) (StateAction[State], error) {
	br := state.br

	reader := bufio.NewReader(os.Stdin)
	actionMapper := map[string]func() error {
		"create": br.SendPoolCreateMessage,
		"join": func() error {
			poolID, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			poolID = poolID[:len(poolID)-1]
			return br.SendPoolJoinMessage(poolID)
		},
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			msg, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			msg = msg[:len(msg)-1]
			action, exists := actionMapper[msg]
			if !exists {
				log.Printf("SYNTAX_ERROR: no such command as %s\n", msg)
				continue
			}
			action()
		}
	}
}
