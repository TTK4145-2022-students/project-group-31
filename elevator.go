package main

import (
	"elevio"
)

const NUM_FLOORS = 4
const NUM_BUTTONS = 3

type ElevatorBehavior int

const (
	EB_Idle     ElevatorBehavior = 0
	EB_DoorOpen ElevatorBehavior = 1
	EB_Moving   ElevatorBehavior = 2
)

type Elevator struct {
	floor            int
	direction        elevio.MotorDirection
	requests         [NUM_FLOORS][NUM_BUTTONS]bool
	behavior         ElevatorBehavior
	doorOpenDuration int
}

func (e Elevator) setAllLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, e.requests[floor][btn])
		}
	}
}
