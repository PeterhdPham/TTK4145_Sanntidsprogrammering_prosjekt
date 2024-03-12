package defs

import "Driver-go/elevio"

type ButtonEventWithIP struct {
	Event elevio.ButtonEvent
	IP    string
}

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour string
}

type LightStatus struct {
	HallLightUp   []bool `json:"HallLightUp"`
	HallLightDown []bool `json:"HallLightDown"`
	CabLight      []bool `json:"CabLight"`
}

type MasterList struct {
	Elevators []Elevator `json:"elevators"`
}

type Elevator struct {
	Ip       string       `json:"ip"`
	IsOnline bool         `json:"online"`
	Status   ElevStatus   `json:"status"`
	Orders   [][]bool     `json:"orders"`
	Role     ElevatorRole `json:"role"`
}

type ElevStatus struct {
	Direction   int    `json:"direction"`
	Floor       int    `json:"floor"`
	Doors       bool   `json:"doors"`
	Obstructed  bool   `json:"obstructed"`
	Buttonfloor int    `json:"buttonfloor"`
	Buttontype  int    `json:"buttontype"`
	FSM_State   string `json:"fsm_state"`
	Operative   bool   `json:"operator`
}

type ElevatorRole int

type IpRequestCount struct {
	Ip       string
	Requests int
}
