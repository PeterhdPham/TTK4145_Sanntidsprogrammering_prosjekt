package elev_data

import (
	"Driver-go/elevio"
	"encoding/json"
)

type Elev_status struct {
	Direction   int  `json:"direction"`
	Floor       int  `json:"floor"`
	Doors       bool `json:"doors"`
	Obstructed  bool `json:"obstructed"`
	Buttonfloor int  `json:"buttonfloor"`
	Buttontype  int  `json:"buttontype"`
}

type Elev_light struct {
	Up   bool `json:"up"`
	Down bool `json:"down"`
	Cab  bool `json:"cab"`
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
			my_status.Buttonfloor = a.Floor
			my_status.Buttontype = int(a.Button)
			elev_status_chan <- my_status

		case a := <-drv_floors:
			my_status.Buttonfloor = -1
			my_status.Buttontype = -1
			my_status.Floor = a
			elev_status_chan <- my_status

		case a := <-drv_obstr:
			my_status.Buttonfloor = -1
			my_status.Buttontype = -1
			if a {
				my_status.Obstructed = true
			} else {
				my_status.Obstructed = false
			}
			elev_status_chan <- my_status

		case a := <-direction:
			my_status.Buttonfloor = -1
			my_status.Buttontype = -1
			my_status.Direction = int(a)

		case a := <-door_open:
			my_status.Buttonfloor = -1
			my_status.Buttontype = -1
			if a {
				my_status.Doors = true
			} else {
				my_status.Doors = false
			}
			elev_status_chan <- my_status
		}
	}
}

func Status_to_bytestream(status_to_send Elev_status) []byte {
	byte_slice, err := json.Marshal(status_to_send)
	if err != nil {
		panic(err)
	}
	return byte_slice
}
