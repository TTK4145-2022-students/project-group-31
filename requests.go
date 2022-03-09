package main

import (
	"elevio"
)

func RequestsAbove(e Elevator) bool {
	for f := e.floor + 1; f < NUM_FLOORS; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}
func RequestsBelow(e Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func RequestsHere(e Elevator) bool {
	for btn := 0; btn < NUM_BUTTONS; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func NextAction(e Elevator) (direction elevio.MotorDirection, behavior ElevatorBehavior) {

	switch e.direction {
	case elevio.MD_Up:
		if RequestsAbove(e) {
			direction = elevio.MD_Up
			behavior = EB_Moving
		} else if RequestsHere(e) {
			direction = elevio.MD_Down
			behavior = EB_DoorOpen
		} else if RequestsBelow(e) {
			direction = elevio.MD_Down
			behavior = EB_Moving
		} else {
			direction = elevio.MD_Stop
			behavior = EB_Idle
		}
	case elevio.MD_Down:
		if RequestsBelow(e) {
			direction = elevio.MD_Down
			behavior = EB_Moving
		} else if RequestsHere(e) {
			direction = elevio.MD_Up
			behavior = EB_DoorOpen
		} else if RequestsAbove(e) {
			direction = elevio.MD_Up
			behavior = EB_Moving
		} else {
			direction = elevio.MD_Stop
			behavior = EB_Idle
		}
	case elevio.MD_Stop:
		if RequestsHere(e) {
			direction = elevio.MD_Stop
			behavior = EB_DoorOpen
		} else if RequestsAbove(e) {
			direction = elevio.MD_Up
			behavior = EB_Moving
		} else if RequestsBelow(e) {
			direction = elevio.MD_Down
			behavior = EB_Moving
		} else {
			direction = elevio.MD_Stop
			behavior = EB_Idle
		}
	default:
		direction = elevio.MD_Stop
		behavior = EB_Idle
	}
	return
}

func ShouldStop(e Elevator) bool {

	switch e.direction {
	case elevio.MD_Down:
		return e.requests[e.floor][elevio.BT_HallDown] ||
			e.requests[e.floor][elevio.BT_Cab] ||
			!RequestsBelow(e) //???
	case elevio.MD_Up:
		return e.requests[e.floor][elevio.BT_HallUp] ||
			e.requests[e.floor][elevio.BT_Cab] ||
			!RequestsAbove(e)
	case elevio.MD_Stop:
	default:
		return false
	}
	return false
}

func ShouldClearImmediately(e Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	return e.floor == btnFloor && (e.direction == elevio.MD_Up && btnType == elevio.BT_HallUp ||
		e.direction == elevio.MD_Down && btnType == elevio.BT_HallDown ||
		e.direction == elevio.MD_Stop ||
		btnType == elevio.BT_Cab)
}

func clearAtCurrentFloor(e *Elevator) {
	e.requests[e.floor][elevio.BT_Cab] = false
	switch e.direction {
	case elevio.MD_Down:
		if !RequestsAbove(*e) && !e.requests[e.floor][elevio.BT_HallUp] {
			e.requests[e.floor][elevio.BT_HallDown] = false
		}
		e.requests[e.floor][elevio.BT_HallUp] = false
	case elevio.MD_Up:
		if !RequestsBelow(*e) && !e.requests[e.floor][elevio.BT_HallDown] {
			e.requests[e.floor][elevio.BT_HallUp] = false
		}
		e.requests[e.floor][elevio.BT_HallDown] = false
	case elevio.MD_Stop:
	default:
		e.requests[e.floor][elevio.BT_HallUp] = false
		e.requests[e.floor][elevio.BT_HallDown] = false
	}
}
