package elevalgo

import (
	"time"
)

var (
	timerEndTime time.Time
	timerActive  bool
	timerChannel chan bool // Channel to signal the timeout event

	failureDeadline     time.Time
	failuretimerActive  bool
	failureTimerChannel chan int
)

func init() {
	// Initialize the timerChannel
	timerChannel = make(chan bool, 1)       // Buffered channel to prevent blocking on send
	failureTimerChannel = make(chan int, 1) // Buffered channel to prevent blocking on send
}

// timerStart starts the timer with a specified duration in seconds.
func timerStart(duration time.Duration) {
	timerEndTime = time.Now().Add(duration)
	timerActive = true
	go func() {
		time.Sleep(duration)
		if timerActive && time.Now().After(timerEndTime) {
			timerChannel <- true // Signal that the timer has expired
		}
	}()
}

func failureTimerStart(duration time.Duration, mode int) {
	failureDeadline = time.Now().Add(duration)
	failuretimerActive = true
	go func() {
		time.Sleep(duration)
		if failuretimerActive && time.Now().After(failureDeadline) {
			failureTimerChannel <- mode // Signal that the timer has expired
		}
	}()
}

func timerStop() {
	timerActive = false
	// Optionally, clear the channel if stopping prematurely
	select {
	case <-timerChannel:
	default:
	}
}

func failureTimerStop() {
	failuretimerActive = false
	// Optionally, clear the channel if stopping prematurely
	select {
	case <-failureTimerChannel:
	default:
	}
}
