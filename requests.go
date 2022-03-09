package main

import (
	"Driver-go/elevio"
)

func RequestsAbove(e Elevator) bool {
	for f := e.Floor + 1; f < NUM_FLOORS; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}
func RequestsBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func RequestsHere(e Elevator) bool {
	for btn := 0; btn < NUM_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func NextAction(e Elevator) (direction elevio.MotorDirection, behavior ElevatorBehavior) {

	switch e.Direction {
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

	switch e.Direction {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] ||
			e.Requests[e.Floor][elevio.BT_Cab] ||
			!RequestsBelow(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] ||
			e.Requests[e.Floor][elevio.BT_Cab] ||
			!RequestsAbove(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func ShouldClearImmediately(e Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	return e.Floor == btnFloor && (e.Direction == elevio.MD_Up && btnType == elevio.BT_HallUp ||
		e.Direction == elevio.MD_Down && btnType == elevio.BT_HallDown ||
		e.Direction == elevio.MD_Stop ||
		btnType == elevio.BT_Cab)
}

func clearAtCurrentFloor(e *Elevator) {
	e.RemoveOrder(e.Floor, elevio.BT_Cab)
	switch e.Direction {
	case elevio.MD_Down:
		if !RequestsBelow(*e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		}
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	case elevio.MD_Up:
		if !RequestsAbove(*e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
		e.Requests[e.Floor][elevio.BT_HallUp] = false
	case elevio.MD_Stop:
		e.Requests[e.Floor][elevio.BT_HallUp] = false
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	default:
		e.Requests[e.Floor][elevio.BT_HallUp] = false
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	}
}
