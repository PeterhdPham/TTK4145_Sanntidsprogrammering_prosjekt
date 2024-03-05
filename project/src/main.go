package main

import (
	"Driver-go/elevio"
	"encoding/json"
	"fmt"
	elevalgo "project/elevAlgo"
	"project/elevData"
	"project/tcp"
	"time"
)

const N_FLOORS int = 4

var elevator elevData.Elevator
var masterElevator []elevData.Elevator

func main() {

	fmt.Println("Booting elevator") // just to know we're running

	elevator = elevData.InitElevator(N_FLOORS)
	masterElevator = append(masterElevator, elevator)

	myStatus := make(chan elevData.ElevStatus)
	myOrders := make(chan [][]bool)
	go elevData.InitOrdersChan(myOrders, N_FLOORS)

	go tcp.Config_Roles(&elevator)

	elevio.Init("localhost:15657", N_FLOORS) // connect to elevatorsimulator

	ticker := time.NewTicker(10 * time.Second)

	time.Sleep(5 * time.Second)

	go elevalgo.ElevAlgo(masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role)

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
				fmt.Println("IM MASTER:", elevator.Status)
				continue
			}
		case newOrders := <-myOrders:
			fmt.Println("New orders: ", newOrders)
			elevator.Orders = newOrders
		case <-ticker.C:
			fmt.Println("Active ips: ", tcp.ActiveIPs)
			// 	byteStream, err := json.Marshal(elevator)
			// 	if err != nil {
			// 		panic(err)
			// 	}

			// 	fmt.Println(string(byteStream))
		}
	}
}
