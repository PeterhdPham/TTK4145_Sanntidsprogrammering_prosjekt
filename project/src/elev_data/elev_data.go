package elev_data

import "Driver-go/elevio"

func Get_livedata2(
	chan1 chan<- elevio.ButtonEvent,
	chan2 chan<- int,
	chan3 chan<- bool,
	chan4 chan<- bool,
) {

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
			chan1 <- a

		case a := <-drv_floors:
			chan2 <- a

		case a := <-drv_obstr:
			chan3 <- a

		case a := <-drv_stop:
			chan4 <- a
		}
	}
}

func Get_livedata(
	elev_status_chan chan<- []int,
	direction chan elevio.MotorDirection,
	door_open chan bool,
) {

	elev_status := []int{0, 0, 0, 0, 0, 0}
	// Slice contains:[direction(-1,0,1), Floor I'm at (0,1,2,3...n), Doors open? (0,1), Door obstructed? (0,1)
	// buttonfloor(0,1,2,3...n), buttontype(0,1,2)]
	// buttontype 0 is up, type 1 is down, type 2 is hall call

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
			elev_status[4] = a.Floor
			elev_status[5] = int(a.Button)
			elev_status_chan <- elev_status

		case a := <-drv_floors:
			elev_status[4] = -1
			elev_status[5] = -1
			elev_status[1] = a
			elev_status_chan <- elev_status

		case a := <-drv_obstr:
			elev_status[4] = -1
			elev_status[5] = -1
			if a {
				elev_status[3] = 1
			} else {
				elev_status[3] = 0
			}
			elev_status_chan <- elev_status

		case a := <-direction:
			elev_status[4] = -1
			elev_status[5] = -1
			elev_status[0] = int(a)

		case a := <-door_open:
			elev_status[4] = -1
			elev_status[5] = -1
			if a {
				elev_status[2] = 1
			} else {
				elev_status[2] = 0
			}
			elev_status_chan <- elev_status
		}
	}
}
