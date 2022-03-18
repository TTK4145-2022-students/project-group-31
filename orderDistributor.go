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
	localElevIDChan <-chan string) {
	for {
		select {
		case btn := <-drv_buttons:
			elevatorNetwork := <-elevatorNetworkChan
			if elevatorNetwork.OnlyLocal {
				newOrderChan <- btn
			} else {
				fmt.Println("Trying to calc cost")
				if btn.Button == elevio.BT_HallUp || btn.Button == elevio.BT_HallDown {

					var costs [MAX_NUMBER_OF_ELEVATORS]int
					for elevID := 0; elevID < MAX_NUMBER_OF_ELEVATORS; elevID++ {
						if elevatorNetwork.ElevatorModules[elevID].Connected {
							costs[elevID] = CalculateCost(elevatorNetwork.ElevatorModules[elevID].Elevator, btn)
						} else {
							costs[elevID] = 1000000
						}
					}

					minCostElevID := 0
					min := costs[0]
					for i, c := range costs {
						/* fmt.Printf("Cost of elevator %#v", i)
						fmt.Printf("is: %#v\n", c) */
						if c < min {
							min = c
							minCostElevID = i
						}
					}
					elev := elevatorNetwork.ElevatorModules[minCostElevID].Elevator
					elev.AddOrder(btn.Floor, btn.Button)
					distributeOrderChan <- ElevatorMessage{strconv.Itoa(minCostElevID), elev}
				} else {
					//SEND to network instead but send to this elev
					//newOrderChan <- btn
					id, _ := strconv.Atoi(<-localElevIDChan)
					elev := elevatorNetwork.ElevatorModules[id].Elevator
					elev.AddOrder(btn.Floor, btn.Button)
					distributeOrderChan <- ElevatorMessage{strconv.Itoa(id), elev}
				}
			}

		case order := <-networkOrder:
			newOrderChan <- order
		}
	}
}
