package main

import (
	"Driver-go/elevio"
	"encoding/json"
	"fmt"
	"project/elevData"
	"project/tcp"
	"time"
)

const N_FLOORS int = 4

func main() {

	fmt.Println("Booting elevator") // just to know we're running

	var elevator = elevData.InitElevator(N_FLOORS)

	go tcp.Config_Roles(&elevator)

	elevio.Init("localhost:15657", N_FLOORS) // connect to elevatorsimulator

	byteStream, err := json.Marshal(elevator)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(byteStream))

	myStatus := make(chan elevData.ElevStatus) // need these for testing
	myDirection := make(chan elevio.MotorDirection)
	myDoor := make(chan bool)

	go elevData.UpdateStatus(myStatus, myDirection, myDoor) // testing this

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case newStatus := <-myStatus:
			fmt.Println("New status: ", newStatus)
		case <-ticker.C:
			// fmt.Println(tcp.ActiveIPs)
			byteStream, err := json.Marshal(elevator.Role)
			if err != nil {
				panic(err)
			}

			fmt.Println(string(byteStream))
		}
	}
}
