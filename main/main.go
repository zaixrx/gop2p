package main

import (
	"os"
	"log"
	"bufio"
	"context"
)

func main() {
	closePingChan := make(chan struct {})
	brStage := CreateBroadcastingStage(closePingChan, func() (BRMsgType, []string) {
		reader := bufio.NewReader(os.Stdin)

		readStr := func() string {
			txt, _ := reader.ReadString('\n')
			txt = txt[:len(txt)-1]
			return txt
		}

		for {
			msg := readStr()
			switch msg {
			case "create":
				return BRCreatePool, nil
			case "join":
				return BRJoinPool, []string{readStr()} 
			default:
				log.Printf("Invalid Message %s", msg)
			}
		}
	})

	brState, termErr := brStage.Run(nil)
	if termErr != nil {
		log.Fatal(termErr)
	}

	p2p := CreateP2PStage(brState.CurrentPool, 0)
	cancelP2PChan := make(chan context.CancelFunc)
	_, termErr = p2p.Run(cancelP2PChan)
	if termErr != nil {
		log.Fatal(termErr)
	}

	cancelP2P := <-cancelP2PChan
	<-closePingChan
	cancelP2P()
}
