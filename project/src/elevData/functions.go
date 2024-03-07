package elevData

import (
	"Driver-go/elevio"
	"encoding/json"
	"project/light_status"
	"project/udp"
)

func InitElevator(NumberOfFloors int) Elevator {
	var elevator Elevator
	ip, _ := udp.GetPrimaryIP()
	elevator.Lights = light_status.InitLights(NumberOfFloors)
	elevator.Status.Buttonfloor = -1
	elevator.Status.Buttontype = -1
	elevator.Ip = ip
	return elevator
}

func InitOrders(NumberOfFloors int) [][]bool {
	orders := make([][]bool, NumberOfFloors)
	for i := range orders {
		orders[i] = make([]bool, 3)
	}
	return orders
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

func StatusToBytestream(statusToSend ElevStatus) []byte {
	byteSlice, err := json.Marshal(statusToSend)
	if err != nil {
		panic(err)
	}
	return byteSlice
}

func BytestreamToStatus(byteSlice []byte) ElevStatus {
	var status ElevStatus
	err := json.Unmarshal(byteSlice, &status)
	if err != nil {
		panic(err)
	}
	return status
}

func UpdateMasterList(masterList *MasterList, newStatus ElevStatus, ip string) {
	for i := 0; i < len(masterList.Elevators); i++ {
		if masterList.Elevators[i].Ip == ip {
			masterList.Elevators[i].Status = newStatus
		}
	}
}
