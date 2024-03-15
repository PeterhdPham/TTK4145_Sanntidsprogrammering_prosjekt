package elevatorAlgorithm

import (
	"time"
)

var (
	doorTimerEndTime time.Time
	doorTimerActive  bool
	doorTimerChannel chan bool

	failureDeadline     time.Time
	failureTimerActive  bool
	failureTimerChannel chan int
)

func init() {

	doorTimerChannel = make(chan bool, 1)
	failureTimerChannel = make(chan int, 1)
}

func doorTimerStart(duration time.Duration) {
	doorTimerEndTime = time.Now().Add(duration)
	doorTimerActive = true
	go func() {
		time.Sleep(duration)
		if doorTimerActive && time.Now().After(doorTimerEndTime) {
			doorTimerChannel <- true
		}
	}()
}

func failureTimerStart(duration time.Duration, mode int) {
	failureDeadline = time.Now().Add(duration)
	failureTimerActive = true
	go func() {
		time.Sleep(duration)
		if failureTimerActive && time.Now().After(failureDeadline) {
			failureTimerChannel <- mode
		}
	}()
}

func doorTimerStop() {
	doorTimerActive = false

	select {
	case <-doorTimerChannel:
	default:
	}
}

func failureTimerStop() {
	failureTimerActive = false

	select {
	case <-failureTimerChannel:
	default:
	}
}
