package elevalgo

import (
	"Driver-go/elevio"
	"fmt"
	"project/elevData"
)

func ElevAlgo(masterList []elevData.Elevator, elevStatus chan elevData.ElevStatus) {
	fmt.Println("ElevAlgo")

	myStatus := <-elevStatus

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
			elevStatus <- myStatus

		case a := <-drvFloors:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			myStatus.Floor = a
			elevStatus <- myStatus

		case a := <-drvObstr:
			myStatus.Buttonfloor = -1
			myStatus.Buttontype = -1
			if a {
				myStatus.Obstructed = true
			} else {
				myStatus.Obstructed = false
			}
			elevStatus <- myStatus

			// case a := <-direction:
			// 	myStatus.Buttonfloor = -1
			// 	myStatus.Buttontype = -1
			// 	myStatus.Direction = int(a)

			// case a := <-doorOpen:
			// 	myStatus.Buttonfloor = -1
			// 	myStatus.Buttontype = -1
			// 	if a {
			// 		myStatus.Doors = true
			// 	} else {
			// 		myStatus.Doors = false
			// 	}
			// 	elevStatus <- myStatus
		}
	}

}
