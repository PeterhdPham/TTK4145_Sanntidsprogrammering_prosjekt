package main

import (
	"fmt"
	"project/pack"
)

func main() {

	living_IPs := make(chan []string)

	go pack.Broadcast_life()
	go pack.Look_for_life(living_IPs)

	for a := range living_IPs {
		fmt.Printf("\033[2J\033[H")
		fmt.Printf("%d living_IPs: %s\n", len(a), a)
	}
}

