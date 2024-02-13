package main

import (
	"project/pack"
	"fmt"
)

func main() {
	fmt.Print("Running main")
	go pack.Broadcast_life()
}