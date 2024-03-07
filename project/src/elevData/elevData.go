package elevData

import (
	"Driver-go/elevio"
	"encoding/json"
	// "fmt"
)

type MasterList struct {
	Elevators []Elevator `json:"elevators"`
}

type Elevator struct {
	Status ElevStatus   `json:"status"`
	Orders []int        `json:"orders"`
	Role   ElevatorRole `json:"role"`
}

type ElevStatus struct {
	Direction   int  `json:"direction"`
	Floor       int  `json:"floor"`
	Doors       bool `json:"doors"`
	Obstructed  bool `json:"obstructed"`
	Buttonfloor int  `json:"buttonfloor"`
	Buttontype  int  `json:"buttontype"`
}

type ElevatorRole int

const (
	Undefined ElevatorRole = -1
	Master    ElevatorRole = 0
	Slave     ElevatorRole = 1
)

func InitElevator(NumberOfFloors int) Elevator {
	var elevator Elevator
	elevator.Status.Buttonfloor = -1
	elevator.Status.Buttontype = -1
	return elevator
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
