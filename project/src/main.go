package main

import (
	"Driver-go/elevio"
	"fmt"
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

	fmt.Println("Booting elevator") // just to know we're running

	elevator = elevData.InitElevator(defs.N_FLOORS)
	masterElevator.Elevators = append(masterElevator.Elevators, elevator)

	myStatus := make(chan defs.ElevStatus)
	myOrders := make(chan [][]bool)
	go elevData.InitOrdersChan(myOrders, defs.N_FLOORS)

	go tcp.Config_Roles(&elevator, &masterElevator)

	defs.MyIP, _ = udp.GetPrimaryIP()

	elevio.Init("localhost:15657", defs.N_FLOORS) // connect to elevatorsimulator

	ticker := time.NewTicker(5 * time.Second)

	// time.Sleep(5 * time.Second)

	go elevalgo.ElevAlgo(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role)

	for {
		select {
		case newStatus := <-myStatus:
			elevator.Status = newStatus

			//Sends message to server
			if tcp.ServerConnection != nil && elevator.Role == defs.SLAVE {
				fmt.Println("Status: ", newStatus)
				byteStream := utility.MarshalJson(newStatus)
				message := []byte(string(byteStream))                                          // Convert message to byte slice
				err := tcp.SendMessage(tcp.ServerConnection, message, reflect.TypeOf(message)) // Assign the error value to "err"
				if err != nil {
					fmt.Printf("Error sending elevator data: %s\n", err)
				}
			} else if elevator.Role == defs.MASTER {
				elevData.UpdateStatusMasterList(&masterElevator, elevator.Status, defs.MyIP)
			}
		case newOrders := <-myOrders:
			if elevator.Role == defs.MASTER {
				elevData.UpdateLightsMasterList(&masterElevator, defs.MyIP)
				elevalgo.SetAllLights(masterElevator)
			}
			if !utility.SlicesAreEqual(elevator.Orders, newOrders) {
				elevator.Orders = newOrders
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
			// fmt.Println("MasterList: ", masterElevator)
			// fmt.Println("Active ips: ", tcp.ActiveIPs)
			continue
		}
	}
}
