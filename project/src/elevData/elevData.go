package elevData

import (
	"project/light_status"
)

type MasterList struct {
	Elevators []Elevator `json:"elevators"`
}

type Elevator struct {
	Ip       string                   `json:"ip"`
	IsOnline bool                     `json:"online"`
	Status   ElevStatus               `json:"status"`
	Lights   light_status.LightStatus `json:"lights"`
	Orders   [][]bool                 `json:"orders"`
	Role     ElevatorRole             `json:"role"`
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
