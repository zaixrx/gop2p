package main

import (
	"os"
	"log"
	"bufio"
	"context"
)

func main() {
	brCloseChan := make(chan struct {})
	brStage := CreateBroadcastingStage(brCloseChan, func() (BRMsgType, []string) {
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

	brStateChan := make(chan *BRState)
	go brStage.Run(brStateChan, nil)
	brState := <-brStateChan

	p2p := CreateP2PStage(brState.CurrentPool)
	cancelP2PChan := make(chan context.CancelFunc)
	go p2p.Run(nil, cancelP2PChan)
	cancelP2P := <- cancelP2PChan

	<-brCloseChan
	cancelP2P()

	bufio.NewReader(os.Stdin).ReadString('\n')
}
