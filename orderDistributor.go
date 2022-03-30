package main

import (
	"Driver-go/elevio"
	"fmt"
	"strconv"
	"time"
)

func orderDistributor(
	localID string,
	drv_buttons <-chan elevio.ButtonEvent,
	elevatorNetworkUpdateCh <-chan [NUM_ELEVATORS]Elevator,
	orderToLocalElevatorCh chan<- elevio.ButtonEvent,
	distributedOrderCh chan<- NetworkMessage) {
	var elevatorNetworkCopy [NUM_ELEVATORS]Elevator
	for {
		select {
		case btn := <-drv_buttons:
			if btn.Button == elevio.BT_Cab {
				id, _ := strconv.Atoi(localID)
				elevator := elevatorNetworkCopy[id]
				elevator.AddOrder(btn)
				distributedOrderCh <- NetworkMessage{SenderID: localID, MessageType: MT_NewOrder, ElevatorID: localID, Elevator: elevator, TimeStamp: time.Now()}
			} else {
				elevatorID, minElevator := findMinCostElevator(elevatorNetworkCopy, btn)
				fmt.Println("Elevator: ", elevatorID, "received the order")
				distributedOrderCh <- NetworkMessage{SenderID: localID, MessageType: MT_NewOrder, ElevatorID: elevatorID, Elevator: minElevator, TimeStamp: time.Now()}
			}
		case elevatorNetwork := <-elevatorNetworkUpdateCh:
			id, _ := strconv.Atoi(localID)
			for floor := 0; floor < NUM_FLOORS; floor++ {
				for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
					if elevatorNetwork[id].Orders[floor][btn] != elevatorNetworkCopy[id].Orders[floor][btn] {
						if elevatorNetwork[id].Orders[floor][btn] {
							orderToLocalElevatorCh <- elevio.ButtonEvent{Floor: floor, Button: btn}
						}
						//Else order has been completed and we do nothin at the moment
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
			clearAtCurrentFloor(&elevator)
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

func findMinCostElevator(elevatorNetwork [NUM_ELEVATORS]Elevator, order elevio.ButtonEvent) (elevatorID string, elevator Elevator) {
	elevatorID = "0"
	minCost := 10000000
	for id := 0; id < NUM_ELEVATORS; id++ {
		cost := CalculateCost(elevatorNetwork[id], order)
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
