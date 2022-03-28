package main

import "Driver-go/elevio"



func CalculateCost(elevator Elevator, newOrder elevio.ButtonEvent) int {
	elevator.AddOrder(newOrder.Floor, newOrder.Button)
	return TimeToIdle(elevator)
}

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
