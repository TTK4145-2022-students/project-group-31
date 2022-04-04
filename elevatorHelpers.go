package main

import (
	"Driver-go/elevio"
)

type ElevatorBehavior int

const (
	EB_Idle        ElevatorBehavior = 0
	EB_DoorOpen    ElevatorBehavior = 1
	EB_Moving      ElevatorBehavior = 2
	EB_Unavailable ElevatorBehavior = 3
	EB_Initialize  ElevatorBehavior = 4
)

type Elevator struct {
	Floor     int
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior
	Orders    [NUM_FLOORS][NUM_BUTTONS]bool
}

func (elevator *Elevator) Initialize() {
	elevator.Floor = 0
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Initialize
	elevator.SetCabLights()
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)
}

func (e *Elevator) AddOrder(order elevio.ButtonEvent) {
	e.Orders[order.Floor][order.Button] = true
}

func (e *Elevator) RemoveOrder(order elevio.ButtonEvent) {
	e.Orders[order.Floor][order.Button] = false
}

func (e *Elevator) clearAtCurrentFloor() {
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

func (e Elevator) SetCabLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, e.Orders[floor][elevio.BT_Cab])
	}
}

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
