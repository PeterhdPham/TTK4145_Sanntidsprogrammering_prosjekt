package main

import (
	"fmt"
	"os"
	"project/pack"
	"strconv"
)

func main() {

	living_IPs := make(chan []string)

	if len(os.Args) == 1 {
		go pack.Broadcast_life()
		select {}
	} else {
		go pack.Look_for_life(living_IPs)
		for {
			select {
			case a := <-living_IPs:
				fmt.Printf("%s living_IPs: %s\n", strconv.Itoa(len(a)), a)
			}
		}
	}
}
