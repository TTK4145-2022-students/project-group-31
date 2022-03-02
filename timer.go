package main

import (
	"time"
)

func TimerStart(timerChannel chan int, dur int) {
	time.Sleep(time.Duration(dur) * time.Second)
	timerChannel <- 1
}
