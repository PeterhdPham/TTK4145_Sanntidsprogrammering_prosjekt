package elevatorAlgorithm

import (
	"time"
)

var (
	doorTimerEndTime time.Time
	doorTimerActive  bool
	doorTimerChannel chan bool // Channel to signal the timeout event

	failureDeadline     time.Time
	failureTimerActive  bool
	failureTimerChannel chan int
)

func init() {
	// Initialize the doorTimerChannel
	doorTimerChannel = make(chan bool, 1)   // Buffered channel to prevent blocking on send
	failureTimerChannel = make(chan int, 1) // Buffered channel to prevent blocking on send
}

// DoorTimerStart starts the timer with a specified duration in seconds.
func doorTimerStart(duration time.Duration) {
	doorTimerEndTime = time.Now().Add(duration)
	doorTimerActive = true
	go func() {
		time.Sleep(duration)
		if doorTimerActive && time.Now().After(doorTimerEndTime) {
			doorTimerChannel <- true // Signal that the timer has expired
		}
	}()
}

func failureTimerStart(duration time.Duration, mode int) {
	failureDeadline = time.Now().Add(duration)
	failureTimerActive = true
	go func() {
		time.Sleep(duration)
		if failureTimerActive && time.Now().After(failureDeadline) {
			failureTimerChannel <- mode // Signal that the timer has expired
		}
	}()
}

func doorTimerStop() {
	doorTimerActive = false
	// Optionally, clear the channel if stopping prematurely
	select {
	case <-doorTimerChannel:
	default:
	}
}

func failureTimerStop() {
	failureTimerActive = false
	// Optionally, clear the channel if stopping prematurely
	select {
	case <-failureTimerChannel:
	default:
	}
}
