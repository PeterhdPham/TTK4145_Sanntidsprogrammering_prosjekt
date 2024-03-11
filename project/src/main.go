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

	ticker := time.NewTicker(5 * time.Second)

	time.Sleep(5 * time.Second)

	go elevAlgo.ElevAlgo(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role, N_FLOORS)

	for {
		select {
		case newStatus := <-myStatus:
			elevator.Status = newStatus
			// fmt.Println("Role: ", elevator.Role)

			//Sends message to server
			if tcp.ServerConnection != nil && elevator.Role == variable.SLAVE {
				fmt.Println("Status: ", newStatus)
				byteStream := utility.MarshalJson(newStatus)
				message := []byte(string(byteStream)) // Convert message to byte slice
				// fmt.Println("Message: ", string(message))
				err := tcp.SendMessage(tcp.ServerConnection, message, reflect.TypeOf(message)) // Assign the error value to "err"
				if err != nil {
					fmt.Printf("Error sending elevator data: %s\n", err)
				}
			} else if elevator.Role == variable.MASTER {
				// TODO: logic for master status update

				elevData.UpdateStatusMasterList(&masterElevator, elevator.Status, variable.MyIP)
				// jsonToSend := utility.MarshalJson(masterElevator)
				// broadcast.BroadcastMessage(nil, jsonToSend)
			}
		case newOrders := <-myOrders:
			if !utility.SlicesAreEqual(elevator.Orders, newOrders) {
				elevator.Orders = newOrders
				if tcp.ServerConnection != nil && elevator.Role == variable.SLAVE {
					// fmt.Println("New orders: ", newOrders)
					byteStream := utility.MarshalJson(elevator)
					message := []byte(string(byteStream)) // Convert message to byte slice
					fmt.Println("Order message: ", string(message))
					err := tcp.SendMessage(tcp.ServerConnection, message, reflect.TypeOf(message)) // Assign the error value to "err"
					if err != nil {
						fmt.Printf("Error sending elevator data: %s\n", err)
					}
				}
			}

			elevAlgo.SetAllLights(elevator.Orders)
			// elevator.Lights = newOrders
		case <-ticker.C:
			// fmt.Println("MasterList: ", masterElevator)
			// fmt.Println("Active ips: ", tcp.ActiveIPs)
			continue
		}
	}
}
