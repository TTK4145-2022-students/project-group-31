package main

import (
	"time"
)

func TimerStart(timerFinishedChannel chan<- int, dur int) {
	time.Sleep(time.Duration(dur) * time.Second)
	timerFinishedChannel <- 1
}
