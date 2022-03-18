package main

import (
	"Driver-go/elevio"
	"fmt"
	"strconv"
)

func OrderDistributor(
	elevatorNetworkChan <-chan ElevatorNetwork,
	drv_buttons <-chan elevio.ButtonEvent,
	distributeOrderChan chan<- ElevatorMessage,
	newOrderChan chan<- elevio.ButtonEvent,
	networkOrder <-chan elevio.ButtonEvent,
	localElevIDChan <-chan string,
	redistributeOrdersChan <-chan ElevatorNetwork) {
	for {
		select {
		case order := <-drv_buttons:
			fmt.Println("order PRESS")
			elevatorNetwork := <-elevatorNetworkChan
			if elevatorNetwork.OnlyLocal {
				newOrderChan <- order
			} else {
				if order.Button == elevio.BT_HallUp || order.Button == elevio.BT_HallDown {

					distributeOrderChan <- findMinCostElevator(elevatorNetwork, order)
				} else {
					//SEND to network instead but send to this elev
					//newOrderChan <- order
					id, _ := strconv.Atoi(<-localElevIDChan)
					elev := elevatorNetwork.ElevatorModules[id].Elevator
					elev.AddOrder(order.Floor, order.Button)
					distributeOrderChan <- ElevatorMessage{strconv.Itoa(id), elev}
				}
			}

		case order := <-networkOrder:
			newOrderChan <- order
		case elevatorNetwork := <-redistributeOrdersChan:
			fmt.Println("Try to redistribute")
			for id := 0; id < MAX_NUMBER_OF_ELEVATORS; id++ {
				if !elevatorNetwork.ElevatorModules[id].Connected {
					fmt.Println("Found discontinued elevator")
					for floor := 0; floor < NUM_FLOORS; floor++ {
						for btn := 0; btn < NUM_BUTTONS-1; btn++ {
							if elevatorNetwork.ElevatorModules[id].Elevator.Requests[floor][btn] {
								fmt.Println("Found discontinued order")
								order := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}
								distributeOrderChan <- findMinCostElevator(elevatorNetwork, order)
							}
						}
					}
				}
			}
		}
	}
}

func findMinCostElevator(elevatorNetwork ElevatorNetwork, order elevio.ButtonEvent) (elevMsg ElevatorMessage) {

	var costs [MAX_NUMBER_OF_ELEVATORS]int
	for elevID := 0; elevID < MAX_NUMBER_OF_ELEVATORS; elevID++ {
		if elevatorNetwork.ElevatorModules[elevID].Connected {
			costs[elevID] = CalculateCost(elevatorNetwork.ElevatorModules[elevID].Elevator, order)
		} else {
			costs[elevID] = 1000000
		}
	}

	minCostElevID := 0
	min := costs[0]
	for i, c := range costs {
		fmt.Printf("Cost of elevator %#v", i)
		fmt.Printf("is: %#v\n", c)
		if c < min {
			min = c
			minCostElevID = i
		}
	}
	elev := elevatorNetwork.ElevatorModules[minCostElevID].Elevator
	elev.AddOrder(order.Floor, order.Button)
	return ElevatorMessage{strconv.Itoa(minCostElevID), elev}
}
