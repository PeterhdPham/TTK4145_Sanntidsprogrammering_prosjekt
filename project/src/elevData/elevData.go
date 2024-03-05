package elevData

import (
	"project/light_status"
)

type MasterList struct {
	Elevators []Elevator `json:"elevators"`
}

type Elevator struct {
	Ip     string                   `json:"ip"`
	Status ElevStatus               `json:"status"`
	Lights light_status.LightStatus `json:"lights"`
	Orders [][]bool                 `json:"orders"`
}

type ElevStatus struct {
	Direction   int  `json:"direction"`
	Floor       int  `json:"floor"`
	Doors       bool `json:"doors"`
	Obstructed  bool `json:"obstructed"`
	Buttonfloor int  `json:"buttonfloor"`
	Buttontype  int  `json:"buttontype"`
}
