package main

import (
	"fmt"
	// "project/pack"
	"project/udp"
)

func main() {

	living_IPs := make(chan []string)

	go udp.Broadcast_life()
	go udp.Look_for_life(living_IPs)

	for a := range living_IPs {
		fmt.Printf("\033[2J\033[H")
		fmt.Printf("%d living_IPs: %s\n", len(a), a)
	}
}

