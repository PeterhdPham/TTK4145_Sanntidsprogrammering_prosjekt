package elevData

import (
	"Driver-go/elevio"
	"project/udp"
	"project/variable"
)

var RemoteStatus ElevStatus

func InitElevator(NumberOfFloors int) Elevator {
	var elevator Elevator
	ip, _ := udp.GetPrimaryIP()
	elevator.Status.Buttonfloor = -1
	elevator.Status.Buttontype = -1
	elevator.Ip = ip
	elevator.Orders = InitOrders(NumberOfFloors)
	elevator.Status.FSM_State = variable.Idle
	return elevator
}

func InitOrders(NumberOfFloors int) [][]bool {
	orders := make([][]bool, NumberOfFloors)
	for i := range orders {
		orders[i] = make([]bool, 3)
	}
	return orders
}

func InitOrdersChan(orders chan [][]bool, numOfFloors int) {
	o := make([][]bool, numOfFloors)
	for i := 0; i < numOfFloors; i++ {
		o[i] = make([]bool, 3) // Assuming 3 buttons per floor.
	}
	// Send the initialized slice of slices through the channel.
	orders <- o
}

func UpdateStatus(
	elevStatusChan chan<- ElevStatus,
	direction chan elevio.MotorDirection,
	doorOpen <-chan bool,
) {
	var myStatus ElevStatus
	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)

	for {
		select {
		case a := <-drvButtons:
			myStatus.Buttonfloor = a.Floor
			myStatus.Buttontype = int(a.Button)
			elevStatusChan <- myStatus

		case a := <-drvFloors:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			myStatus.Floor = a
			elevStatusChan <- myStatus

		case a := <-drvObstr:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Obstructed = true
			} else {
				myStatus.Obstructed = false
			}
			elevStatusChan <- myStatus

		case a := <-direction:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			myStatus.Direction = int(a)

		case a := <-doorOpen:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Doors = true
			} else {
				myStatus.Doors = false
			}
			elevStatusChan <- myStatus
		}
	}
}

func UpdateMasterList(masterList *MasterList, newStatus ElevStatus, ip string) {
	for i := 0; i < len(masterList.Elevators); i++ {
		if masterList.Elevators[i].Ip == ip {
			masterList.Elevators[i].Status = newStatus
		}
	}
}
