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

	myStatus := make(chan elevData.ElevStatus) // need these for testing
	myDirection := make(chan elevio.MotorDirection)
	myDoor := make(chan bool)

	go elevData.UpdateStatus(myStatus, myDirection, myDoor) // testing this

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case newStatus := <-myStatus:
			fmt.Println("New status: ", newStatus)
			elevator.Status = newStatus

			//Turns data into string
			byteStream, err := json.Marshal(elevator.Status)
			if err != nil {
				panic(err)
			}
			message := string(byteStream)

			//Sends message to server
			if tcp.ServerConnection != nil && elevator.Role == elevData.Slave {
				err = tcp.SendMessage(tcp.ServerConnection, message)
				if err != nil {
					fmt.Printf("Error sending elevator data: %s\n", err)
				}
			} else if elevator.Role == elevData.Master {
				// TODO: logic for master status update
				continue
			}

		case <-ticker.C:
			fmt.Println(tcp.ActiveIPs)
			// 	byteStream, err := json.Marshal(elevator)
			// 	if err != nil {
			// 		panic(err)
			// 	}

			// 	fmt.Println(string(byteStream))
		}
	}
}
