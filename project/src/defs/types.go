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

type MasterList struct {
	Elevators []Elevator
}

type Elevator struct {
	Ip       string
	IsOnline bool
	Status   ElevStatus
	Orders   [][]bool
	Lights   [][]bool
	Role     ElevatorRole
}

type ElevStatus struct {
	Direction   int
	Floor       int
	Doors       bool
	Obstructed  bool
	Buttonfloor int
	Buttontype  int
	FSM_State   string
	Operative   bool
}

type ElevatorRole int

type IpRequestCount struct {
	Ip       string
	Requests int
}

type FailureMode int
