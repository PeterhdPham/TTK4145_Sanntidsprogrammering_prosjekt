package main

import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_order := make(chan elevio.ButtonEvent)

	// elevio.ButtonEvent floor_order = {1,2}


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			fmt.Printf("Button received")
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			elevio.MoveTowardsFloor(a.Floor)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			var floor_order = <- drv_order
			if a == floor_order.Floor {
				d = elevio.MD_Stop
				fmt.Printf("Stop")
				elevio.SetButtonLamp(floor_order.Button, floor_order.Floor, false)
			} 
			// if a == numFloors-1 {
			// 	d = elevio.MD_Down
			// } else if a == 0 {
			// 	d = elevio.MD_Up
			// }
			fmt.Println("Floor")
			// elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			fmt.Printf("Obstruction")
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			fmt.Printf("Stop")
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}


	}
}
