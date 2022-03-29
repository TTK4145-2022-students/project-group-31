package main

import "Driver-go/elevio"

type ElevatorBehavior int

const (
	EB_Idle        ElevatorBehavior = 0
	EB_DoorOpen    ElevatorBehavior = 1
	EB_Moving      ElevatorBehavior = 2
	EB_Unavailable ElevatorBehavior = 3
)

type Elevator struct {
	Floor     int
	Direction elevio.MotorDirection
	Requests  [NUM_FLOORS][NUM_BUTTONS]bool
	Behavior  ElevatorBehavior
	Online    bool
}

func ElevatorFSM(
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	orderToLocalElevatorCh <-chan elevio.ButtonEvent,
	initialElevator chan<- Elevator,
	elevatorStateChangeCh chan<- Elevator) {
}
