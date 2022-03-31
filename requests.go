package main

import (
	"Driver-go/elevio"
)

func OrdersAbove(e Elevator) bool {
	for f := e.Floor + 1; f < NUM_FLOORS; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.Orders[f][btn] {
				return true
			}
		}
	}
	return false
}
func OrdersBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.Orders[f][btn] {
				return true
			}
		}
	}
	return false
}

func OrdersHere(e Elevator) bool {
	for btn := 0; btn < NUM_BUTTONS; btn++ {
		if e.Orders[e.Floor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirection(e Elevator) (direction elevio.MotorDirection, behavior ElevatorBehavior) {
	switch e.Direction {
	case elevio.MD_Up:
		if OrdersAbove(e) {
			direction = elevio.MD_Up
			behavior = EB_Moving
		} else if OrdersHere(e) {
			direction = elevio.MD_Down
			behavior = EB_DoorOpen
		} else if OrdersBelow(e) {
			direction = elevio.MD_Down
			behavior = EB_Moving
		} else {
			direction = elevio.MD_Stop
			behavior = EB_Idle
		}
	case elevio.MD_Down:
		if OrdersBelow(e) {
			direction = elevio.MD_Down
			behavior = EB_Moving
		} else if OrdersHere(e) {
			direction = elevio.MD_Up
			behavior = EB_DoorOpen
		} else if OrdersAbove(e) {
			direction = elevio.MD_Up
			behavior = EB_Moving
		} else {
			direction = elevio.MD_Stop
			behavior = EB_Idle
		}
	case elevio.MD_Stop:
		if OrdersHere(e) {
			direction = elevio.MD_Stop
			behavior = EB_DoorOpen
		} else if OrdersAbove(e) {
			direction = elevio.MD_Up
			behavior = EB_Moving
		} else if OrdersBelow(e) {
			direction = elevio.MD_Down
			behavior = EB_Moving
		} else {
			direction = elevio.MD_Stop
			behavior = EB_Idle
		}
	}
	return
}

func ShouldStop(e Elevator) bool {

	switch e.Direction {
	case elevio.MD_Down:
		return e.Orders[e.Floor][elevio.BT_HallDown] ||
			e.Orders[e.Floor][elevio.BT_Cab] ||
			!OrdersBelow(e)
	case elevio.MD_Up:
		return e.Orders[e.Floor][elevio.BT_HallUp] ||
			e.Orders[e.Floor][elevio.BT_Cab] ||
			!OrdersAbove(e)
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
	e.Orders[e.Floor][elevio.BT_Cab] = false
	switch e.Direction {
	case elevio.MD_Up:
		if !OrdersAbove(*e) && !e.Orders[e.Floor][elevio.BT_HallUp] {
			e.Orders[e.Floor][elevio.BT_HallDown] = false
		}
		e.Orders[e.Floor][elevio.BT_HallUp] = false

	case elevio.MD_Down:
		if !OrdersBelow(*e) && !e.Orders[e.Floor][elevio.BT_HallDown] {
			e.Orders[e.Floor][elevio.BT_HallUp] = false
		}
		e.Orders[e.Floor][elevio.BT_HallDown] = false
	default:
		e.Orders[e.Floor][elevio.BT_HallUp] = false
		e.Orders[e.Floor][elevio.BT_HallDown] = false
	}
}
