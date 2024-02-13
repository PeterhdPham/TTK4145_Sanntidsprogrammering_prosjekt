package pack

import (
	"Driver-go/elevio"
	"fmt"
)

func Test_heis() {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	OpenDoor(d)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			fmt.Println("Button received")
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			TurnOffLight(a.Floor, -1)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
				elevio.SetMotorDirection(d)
			} else if a == 0 {
				d = elevio.MD_Up
				elevio.SetMotorDirection(d)
			}
			elevio.SetFloorIndicator(a)
			OpenDoor(d)
			TurnOffLight(a, d)

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
