package main

import (
	"encoding/json"
	"fmt"
	"project/elevData"
	"project/udp"
)

const N_FLOORS int = 4

func main() {

	livingIPs := make(chan []string)

	go udp.BroadcastLife()
	go udp.LookForLife(livingIPs)

	var masterList elevData.MasterList

	elevator := elevData.InitElevator(N_FLOORS)

	masterList.Elevators = append(masterList.Elevators, elevator)

	bytes, err := json.Marshal(masterList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))

	for a := range livingIPs {
		fmt.Println("Living IPs: ", a)
	}
}
