package main

import (
	"Driver-go/elevio"
	"fmt"
	"project/broadcast"
	"project/defs"
	elevalgo "project/elevAlgo"
	"project/elevData"
	"project/tcp"
	"project/udp"
	"project/utility"
	"reflect"
	"time"
)

var elevator defs.Elevator
var masterElevator defs.MasterList

func main() {
	elevio.Init("localhost:15657", defs.N_FLOORS)   // connect to elevatorsimulator
	elevator = elevData.InitElevator(defs.N_FLOORS) // initialize the elevator

	masterElevator.Elevators = append(masterElevator.Elevators, elevator) // append the elevator to the master list of elevators

	myStatus := make(chan defs.ElevStatus)              // channel to receive status updates
	myOrders := make(chan [][]bool)                     // channel to receive order updates
	go elevData.InitOrdersChan(myOrders, defs.N_FLOORS) // initialize the orders channel

	go tcp.Config_Roles(&elevator, &masterElevator) // initialize the server and client connections

	defs.MyIP, _ = udp.GetPrimaryIP()

	ticker := time.NewTicker(5 * time.Second)

	go elevalgo.ElevAlgo(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role) // initialize the elevator algorithm

	for {
		select {
		case newStatus := <-myStatus:

			//Sends message to server
			if tcp.ServerConnection != nil && elevator.Role == defs.SLAVE {
				if !reflect.DeepEqual(elevator.Status, newStatus) {
					elevator.Status = newStatus
					byteStream := utility.MarshalJson(newStatus)
					message := []byte(string(byteStream))                                          // Convert message to byte slice
					err := tcp.SendMessage(tcp.ServerConnection, message, reflect.TypeOf(message)) // Assign the error value to "err"
					if err != nil {
						fmt.Printf("Error sending elevator data: %s\n", err)
					}
				}
			} else if elevator.Role == defs.MASTER {
				elevator.Status = newStatus
				elevData.UpdateStatusMasterList(&masterElevator, elevator.Status, defs.MyIP)
				byteStream := utility.MarshalJson(masterElevator)
				broadcast.BroadcastMessage(nil, byteStream)
			}
			elevalgo.SetAllLights(masterElevator)

		case newOrders := <-myOrders:
			if elevator.Role == defs.MASTER {
				elevData.UpdateLightsMasterList(&masterElevator, defs.MyIP)
				elevalgo.SetAllLights(masterElevator)
			}
			if !utility.SlicesAreEqual(elevator.Orders, newOrders) {
				elevator.Orders = newOrders
				fmt.Println("Orders: ", newOrders)
				if tcp.ServerConnection != nil && elevator.Role == defs.SLAVE {
					byteStream := utility.MarshalJson(elevator)
					message := []byte(string(byteStream))                                          // Convert message to byte slice
					err := tcp.SendMessage(tcp.ServerConnection, message, reflect.TypeOf(message)) // Assign the error value to "err"
					if err != nil {
						fmt.Printf("Error sending elevator data: %s\n", err)
					}
				}
			}
		case <-ticker.C:
			fmt.Println("MasterList: ", myOrders)
			// fmt.Println("Active ips: ", tcp.ActiveIPs)
			continue
		}
	}
}
