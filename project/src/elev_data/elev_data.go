package elev_data

import "Driver-go/elevio"

func get_livedata(
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

	select{}
}