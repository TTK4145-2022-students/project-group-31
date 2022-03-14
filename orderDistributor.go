package main

import (
	"Driver-go/elevio"
	"strconv"
)

func OrderDistributor(
	elevatorNetworkChan <-chan ElevatorNetwork,
	drv_buttons <-chan elevio.ButtonEvent,
	elevatorMessageChan chan<- ElevatorMessage,
	newOrderChan chan<- elevio.ButtonEvent) {
	for {
		select {
		case btn := <-drv_buttons:
			elevatorNetwork := <-elevatorNetworkChan
			if btn.Button == elevio.BT_HallUp || btn.Button == elevio.BT_HallDown {

				var costs [MAX_NUMBER_OF_ELEVATORS]int
				for elevID := 0; elevID < MAX_NUMBER_OF_ELEVATORS; elevID++ {
					if elevatorNetwork.ElevatorModules[elevID].Connected {
						costs[elevID] = CalculateCost(elevatorNetwork.ElevatorModules[elevID].Elevator, btn)
					} else {
						costs[elevID] = 1000000
					}

				}
				var minCostElevID int
				min := costs[0]
				for i, c := range costs {
					if c < min {
						min = c
						minCostElevID = i
					}
				}
				elev := elevatorNetwork.ElevatorModules[minCostElevID].Elevator
				elev.AddOrder(btn.Floor, btn.Button)
				elevatorMessageChan <- ElevatorMessage{strconv.Itoa(minCostElevID), elev}
			} else {
				newOrderChan <- btn
			}

		}
	}
}
