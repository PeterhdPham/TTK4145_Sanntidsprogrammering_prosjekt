package main

import (
	"Driver-go/elevio"
	"log"
	"project/aliveMessages"
	"project/communication"
	"project/defs"
	"project/elevatorAlgorithm"
	"project/elevatorData"
	"project/roleConfiguration"
	"project/utility"
	"reflect"
	"time"
)

var elevator defs.Elevator
var masterElevator defs.MasterList

func main() {
	elevio.Init("localhost:15657", defs.N_FLOORS) // connect to elevatorsimulator
	elevator = elevatorData.InitElevator()        // initialize the elevator

	masterElevator.Elevators = append(masterElevator.Elevators, elevator) // append the elevator to the master list of elevators

	myStatus := make(chan defs.ElevStatus) // channel to receive status updates
	myOrders := make(chan [][]bool)        // channel to receive order updates
	elevatorData.InitOrdersChan(myOrders)  // initialize the orders channel

	defs.MyIP = aliveMessages.GetPrimaryIP()

	ticker := time.NewTicker(5 * time.Second)

	go roleConfiguration.Config_Roles(&elevator, &masterElevator) // initialize the server and client connections

	go elevatorAlgorithm.ElevatorControlLoop(&masterElevator, myStatus, myOrders, elevator.Orders, elevator.Role) // initialize the elevator algorithm

	for {
		select {
		case newStatus := <-myStatus:
			// log.Println("status update: ", string(utility.MarshalJson(newStatus)))

			//Sends message to server
			if roleConfiguration.ServerConnection != nil && elevator.Role == defs.SLAVE {
				if !reflect.DeepEqual(elevator.Status, newStatus) {
					elevator.Status = newStatus
					// Convert message to byte slice
					err := communication.SendMessage(roleConfiguration.ServerConnection, newStatus, "") // Assign the error value to "err"
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			} else if elevator.Role == defs.MASTER {
				elevator.Status = newStatus
				elevatorData.UpdateStatusMasterList(&masterElevator, elevator.Status, defs.MyIP)
				communication.BroadcastMessage(nil, &masterElevator)
			}
			elevatorData.SetAllLights(masterElevator)

		case newOrders := <-myOrders:
			if elevator.Role == defs.MASTER {
				elevatorData.UpdateLightsMasterList(&masterElevator, defs.MyIP)
				elevatorData.SetAllLights(masterElevator)
			}
			if !utility.SlicesAreEqual(elevator.Orders, newOrders) {
				elevator.Orders = newOrders
				if roleConfiguration.ServerConnection != nil && elevator.Role == defs.SLAVE {
					// Convert message to byte slice
					err := communication.SendMessage(roleConfiguration.ServerConnection, elevator, "") // Assign the error value to "err"
					if err != nil {
						log.Printf("Error sending elevator data: %s\n", err)
					}
				}
			}
		case <-ticker.C:
			bytes := utility.MarshalJson(masterElevator)
			log.Println("")
			log.Println("MasterList: ", string(bytes))
			log.Println("\nActive ips: ", roleConfiguration.ActiveIPs)
			currentIP := aliveMessages.GetPrimaryIP()
			if defs.MyIP != currentIP && currentIP != "" {
				defs.MyIP = currentIP
				for index := range masterElevator.Elevators {
					masterElevator.Elevators[index].Ip = defs.MyIP
				}
			}
			continue
		}
	}
}
