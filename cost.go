package main

import "Driver-go/elevio"



func CalculateCost(elevator Elevator, newOrder elevio.ButtonEvent) int {
	elevator.AddOrder(newOrder.Floor, newOrder.Button)
	return TimeToIdle(elevator)
}

/*func ChooseDirection(elevator Elevator) elevio.MotorDirection {
	switch elevator.Direction {
	case elevio.MD_Up:
		if RequestsAbove(elevator) {
			return elevio.MD_Up
		} else if RequestsHere(elevator) {
			return elevio.MD_Stop
		} else if RequestsBelow(elevator) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}
	case elevio.MD_Down:
		if RequestsBelow(elevator) {
			return elevio.MD_Down
		} else if RequestsHere(elevator) {
			return elevio.MD_Stop
		} else if RequestsAbove(elevator) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	case elevio.MD_Stop:
		if RequestsBelow(elevator) {
			return elevio.MD_Down
		} else if RequestsHere(elevator) {
			return elevio.MD_Stop
		} else if RequestsAbove(elevator) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	}
	//Will never be called
	return elevio.MD_Stop
}*/

func TimeToIdle(elevator Elevator) int {
	duration := 0
	switch elevator.Behavior {
	case EB_Idle:
		elevator.Direction, _ = NextAction(elevator)
		if elevator.Direction == elevio.MD_Stop {
			return duration
		}
	case EB_Moving:
		duration += ELEVATOR_TRAVEL_COST / 2
		elevator.Floor += int(elevator.Direction)
	case EB_DoorOpen:
		duration -= ELEVATOR_DOOR_OPEN_COST / 2
	}
	for {
		if ShouldStop(elevator) {
			clearAtCurrentFloor(&elevator)
			duration += ELEVATOR_DOOR_OPEN_COST
			elevator.Direction, _ = NextAction(elevator)
			if elevator.Direction == elevio.MD_Stop {
				return duration
			}
		}
		elevator.Floor += int(elevator.Direction)
		duration += ELEVATOR_TRAVEL_COST
	}
}
