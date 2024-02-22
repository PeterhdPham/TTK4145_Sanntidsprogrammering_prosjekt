package elev_data

import "Driver-go/elevio"

type Elev_status struct {
	direction   int  `json:"direction"`
	floor       int  `json:"floor"`
	doors       bool `json:"doors"`
	obstructed  bool `json:"obstructed"`
	buttonfloor int  `json:"buttonfloor"`
	buttontype  int  `json:"buttontype"`
}

func Get_livedata(
	elev_status_chan chan<- Elev_status,
	direction chan elevio.MotorDirection,
	door_open chan bool,
) {

	var my_status Elev_status
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			my_status.buttonfloor = a.Floor
			my_status.buttontype = int(a.Button)
			elev_status_chan <- my_status

		case a := <-drv_floors:
			my_status.buttonfloor = -1
			my_status.buttontype = -1
			my_status.floor = a
			elev_status_chan <- my_status

		case a := <-drv_obstr:
			my_status.buttonfloor = -1
			my_status.buttontype = -1
			if a {
				my_status.obstructed = true
			} else {
				my_status.obstructed = false
			}
			elev_status_chan <- my_status

		case a := <-direction:
			my_status.buttonfloor = -1
			my_status.buttontype = -1
			my_status.direction = int(a)

		case a := <-door_open:
			my_status.buttonfloor = -1
			my_status.buttontype = -1
			if a {
				my_status.doors = true
			} else {
				my_status.doors = false
			}
			elev_status_chan <- my_status
		}
	}
}
