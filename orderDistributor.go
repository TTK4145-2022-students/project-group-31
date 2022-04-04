package main

import (
	"Driver-go/elevio"
	"strconv"
)

func orderDistributor(
	localID string,
	drv_buttons <-chan elevio.ButtonEvent,
	elevatorNetworkChangeCh <-chan [NUM_ELEVATORS]Elevator,
	addLocalOrder chan<- elevio.ButtonEvent,
	distributedOrderCh chan<- NetworkMessage) {
	var elevatorNetworkCopy [NUM_ELEVATORS]Elevator
	for {
		select {
		case btn := <-drv_buttons:
			if btn.Button == elevio.BT_Cab {
				id, _ := strconv.Atoi(localID)
				elevator := elevatorNetworkCopy[id]
				elevator.AddOrder(btn)
				distributedOrderCh <- NetworkMessage{
					SenderID:    localID,
					MessageType: MT_NewOrder,
					ElevatorID:  localID,
					Elevator:    elevator}
			} else {
				elevatorID, minElevator := findMinCostElevator(elevatorNetworkCopy, btn)
				if minElevator.Behavior != EB_Unavailable {
					distributedOrderCh <- NetworkMessage{
						SenderID:    localID,
						MessageType: MT_NewOrder,
						ElevatorID:  elevatorID,
						Elevator:    minElevator}
				}
			}
		case elevatorNetwork := <-elevatorNetworkChangeCh:

			for id := 0; id < NUM_ELEVATORS; id++ {
				if elevatorNetwork[id].Behavior == EB_Unavailable && elevatorNetworkCopy[id].Behavior != EB_Unavailable {
					elevatorNetwork = redistributeHallOrders(elevatorNetwork, id)
				}
			}

			id, _ := strconv.Atoi(localID)
			for floor := 0; floor < NUM_FLOORS; floor++ {
				for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
					if elevatorNetwork[id].Orders[floor][btn] != elevatorNetworkCopy[id].Orders[floor][btn] {
						if elevatorNetwork[id].Orders[floor][btn] {
							addLocalOrder <- elevio.ButtonEvent{Floor: floor, Button: btn}
						}
					}
				}
			}
			elevatorNetworkCopy = elevatorNetwork
		}
	}
}

func CalculateCost(elevator Elevator, newOrder elevio.ButtonEvent) int {
	elevator.AddOrder(newOrder)
	duration := 0
	switch elevator.Behavior {
	case EB_Idle:
		elevator.Direction, _ = ChooseDirection(elevator)
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
			elevator.clearAtCurrentFloor()
			duration += ELEVATOR_DOOR_OPEN_COST
			elevator.Direction, _ = ChooseDirection(elevator)
			if elevator.Direction == elevio.MD_Stop {
				return duration
			}
		}
		elevator.Floor += int(elevator.Direction)
		duration += ELEVATOR_TRAVEL_COST
	}
}

func findMinCostElevator(
	elevatorNetwork [NUM_ELEVATORS]Elevator,
	order elevio.ButtonEvent) (
	elevatorID string,
	elevator Elevator) {

	minCost := MAX_COST
	for id := 0; id < NUM_ELEVATORS; id++ {
		cost := MAX_COST
		if elevatorNetwork[id].Behavior == EB_Unavailable {
		} else {
			cost = CalculateCost(elevatorNetwork[id], order)
		}

		if cost < minCost {
			minCost = cost
			elevatorID = strconv.Itoa(id)
		}
	}
	id, _ := strconv.Atoi(elevatorID)
	elevator = elevatorNetwork[id]
	elevator.AddOrder(order)
	return
}

func redistributeHallOrders(elevatorNetwork [NUM_ELEVATORS]Elevator, unavailableID int) [NUM_ELEVATORS]Elevator {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < NUM_BUTTONS-1; btn++ {
			order := elevio.ButtonEvent{Floor: floor, Button: btn}
			if elevatorNetwork[unavailableID].Orders[floor][btn] {
				elevatorID, minElevator := findMinCostElevator(elevatorNetwork, order)
				id, _ := strconv.Atoi(elevatorID)
				elevatorNetwork[id] = minElevator
				elevatorNetwork[unavailableID].RemoveOrder(order)
			}
		}
	}
	return elevatorNetwork
}
