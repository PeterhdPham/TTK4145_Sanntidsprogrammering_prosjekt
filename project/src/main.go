package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevAlgo"
	"project/elevData"
	"project/ip"
	"project/tcp"
	"project/utility"
	"project/variable"
	"reflect"
	"time"
)

const N_FLOORS int = 4

var elevator variable.Elevator
var masterElevator variable.MasterList

func main() {

	fmt.Println("Booting elevator") // just to know we're running

	elevator = elevData.InitElevator(N_FLOORS)
	masterElevator.Elevators = append(masterElevator.Elevators, elevator)

	myStatus := make(chan variable.ElevStatus)
	myOrders := make(chan [][]bool)
	go elevData.InitOrdersChan(myOrders, N_FLOORS)

	go tcp.Config_Roles(&elevator, &masterElevator)

	variable.MyIP, _ = ip.GetPrimaryIP()

	elevio.Init("localhost:15657", N_FLOORS) // connect to elevatorsimulator

	ticker := time.NewTicker(10 * time.Second)

	time.Sleep(5 * time.Second)

	go elevAlgo.ElevAlgo(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role, N_FLOORS)

	for {
		select {
		case newStatus := <-myStatus:

			elevator.Status = newStatus

			//Turns data into string
			byteStream := utility.MarshalJson(elevator.Status)

			message := []byte(string(byteStream)) // Convert message to byte slice

			//Sends message to server
			fmt.Println("Role: ", elevator.Role)
			fmt.Println("Status: ", elevator.Status)
			if tcp.ServerConnection != nil && elevator.Role == variable.SLAVE {
				fmt.Println("Message: ", string(message))
				err := tcp.SendMessage(tcp.ServerConnection, message, reflect.TypeOf(message)) // Assign the error value to "err"
				if err != nil {
					fmt.Printf("Error sending elevator data: %s\n", err)
				}
			} else if elevator.Role == variable.MASTER {
				// TODO: logic for master status update

				elevData.UpdateMasterList(&masterElevator, elevator.Status, variable.MyIP)
				// jsonToSend := utility.MarshalJson(masterElevator)
				// broadcast.BroadcastMessage(nil, jsonToSend)
			}
		case newOrders := <-myOrders:
			// fmt.Println("New orders: ", newOrders)
			elevator.Orders = newOrders
			elevAlgo.SetAllLights(elevator.Orders)
			// elevator.Lights = newOrders
		case <-ticker.C:
			// fmt.Println("Active ips: ", tcp.ActiveIPs)
			continue
		}
	}
}
