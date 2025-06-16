package main

import (
	"log"
	"p2p/main/broadcast"
)

func main() {
	br := Broadcast.CreateBroadcast()
	go br.Start()

	pools, err := br.RetrievePools()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Available Pools", pools)

	pool, err := br.SendCreatePool()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created Pool:", pool)

	// currentPool, err := br.SendCreatePool()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("Current Pool", currentPool)
}	
