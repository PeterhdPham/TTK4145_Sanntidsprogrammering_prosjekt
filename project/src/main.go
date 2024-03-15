package main

import (
	"Driver-go/elevio"
	"log"
	"project/communication"
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

	defs.MyIP, defs.MyPort, _ = udp.GetPrimaryIP()

	ticker := time.NewTicker(5 * time.Second)

	go tcp.Config_Roles(&elevator, &masterElevator) // initialize the server and client connections

	go elevalgo.ElevAlgo(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role) // initialize the elevator algorithm

	for {
		select {
		case newStatus := <-myStatus:
			// log.Println("status update: ", string(utility.MarshalJson(newStatus)))

			//Sends message to server
			if tcp.ServerConnection != nil && elevator.Role == defs.SLAVE {
				if !reflect.DeepEqual(elevator.Status, newStatus) {
					elevator.Status = newStatus
					// Convert message to byte slice
					err := communication.SendMessage(tcp.ServerConnection, newStatus, "") // Assign the error value to "err"
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			} else if elevator.Role == defs.MASTER {
				elevator.Status = newStatus
				elevData.UpdateStatusMasterList(&masterElevator, elevator.Status, defs.MyIP)
				communication.BroadcastMessage(nil, &masterElevator)
			}
			elevData.SetAllLights(masterElevator)

		case newOrders := <-myOrders:
			if elevator.Role == defs.MASTER {
				elevData.UpdateLightsMasterList(&masterElevator, defs.MyIP)
				elevData.SetAllLights(masterElevator)
			}
			if !utility.SlicesAreEqual(elevator.Orders, newOrders) {
				elevator.Orders = newOrders
				log.Println("Orders: ", newOrders)
				if tcp.ServerConnection != nil && elevator.Role == defs.SLAVE {
					// Convert message to byte slice
					err := communication.SendMessage(tcp.ServerConnection, elevator, "") // Assign the error value to "err"
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			}
		case <-ticker.C:
			bytes := utility.MarshalJson(masterElevator)
			log.Println("")
			log.Println("MasterList: ", string(bytes))
			// log.Println("Active ips: ", tcp.ActiveIPs)
			currentIP, _, _ := udp.GetPrimaryIP()
			if defs.MyIP != currentIP && currentIP != ""{
				defs.MyIP = currentIP
				log.Println("Switching ips: ", defs.MyIP)
				for index := range masterElevator.Elevators {
					masterElevator.Elevators[index].Ip = defs.MyIP
				}
			}
			continue
		}
	}
}
