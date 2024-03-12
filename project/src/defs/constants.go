package defs

var N_FLOORS = 4

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
