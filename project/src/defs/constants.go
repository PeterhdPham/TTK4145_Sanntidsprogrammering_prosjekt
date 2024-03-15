package defs

import "time"

const N_FLOORS = 4                                             // Number of floors in the building
const N_BUTTONS = 3                                            // Number of buttons on each floor
const DOOR_OPEN_DURATION time.Duration = 3 * time.Second       // Time to wait before closing door
const FAILURE_TIMEOUT_DURATION time.Duration = 7 * time.Second // Time to wait before declaring a failure
const UDP_PORT = "9999"                                        // UDP_Port used to broadcast and listen to "I'm alive"-messages
const BROADCAST_ADDR = "255.255.255.255:" + UDP_PORT           // Address to broadcast "I'm alive"-msg
const BROADCAST_PERIOD = 100 * time.Millisecond                // Time to wait before broadcasting new msg
const LISTEN_ADDR = "0.0.0.0:" + UDP_PORT                      // Address to listen for "I'm alive"-msg
const LISTEN_TIMEOUT = 10 * time.Second                        // Time to listen before giving up
const NODE_LIFE = 5 * time.Second                              // Time added to node-lifetime when msg is received

const (
	IDLE      = "EB_Idle"
	MOVING    = "EB_Moving"
	DOOR_OPEN = "EB_DoorOpen"
)

const (
	UNDEFINED ElevatorRole = -1
	MASTER    ElevatorRole = 0
	SLAVE     ElevatorRole = 1
)

const (
	DOOR_STUCK FailureMode = 0
	MOTOR_FAIL FailureMode = 1
)
