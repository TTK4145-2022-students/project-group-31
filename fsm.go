package main

import (
	"elevio"
)

var elevator *Elevator

func InitializeElevator() {
	elevator.floor = -1 //Where we are
	elevator.direction = elevio.MD_Down
	elevator.behavior = EB_Idle
	elevator.doorOpenDuration = 3
	elevio.SetMotorDirection(elevator.direction)
}

func onRequestButtonPressed(btnFloor int, btnType elevio.ButtonType) {
	switch elevator.behavior {
	case EB_DoorOpen:
		if ShouldClearImmediately(*elevator, btnFloor, btnType) {
			//START TIMER
		} else {
			elevator.requests[btnFloor][btnType] = true
		}
	case EB_Moving:
		elevator.requests[btnFloor][btnType] = true
	case EB_Idle:
		elevator.requests[btnFloor][btnType] = true
		elevator.direction, elevator.behavior = NextAction(*elevator)
		switch elevator.behavior {
		case EB_DoorOpen:
			//OPEN DOOR AND START TIMER
			clearAtCurrentFloor(elevator)
		case EB_Moving:
			elevio.SetMotorDirection(elevator.direction)
		case EB_Idle:
		}
	}
	elevator.setAllLights()
}

func onFloorArrival(newFloor int) {
	elevator.floor = newFloor
	elevio.SetFloorIndicator(elevator.floor)
	switch elevator.behavior {
	case EB_Moving:
		if ShouldStop(*elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			//Opendoor and Start timer
			clearAtCurrentFloor(elevator)
			elevator.setAllLights()
			elevator.behavior = EB_DoorOpen
		}
	default:
	}
}
