package main

import (
	"os"
	"project/pack"
)

func main() {
	if len(os.Args) == 1 {
		go pack.Broadcast_life()
		select {}
	} else {
		go pack.Look_for_life()
		select {}
	}
}
