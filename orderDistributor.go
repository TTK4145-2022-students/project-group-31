package main

import "Driver-go/elevio"

func orderDistributor(
	id string,
	drv_buttons <-chan elevio.ButtonEvent,
	elevatorNetworkUpdateCh <-chan [NUM_ELEVATORS]Elevator,
	orderToLocalElevatorCh chan<- elevio.ButtonEvent,
	distributedOrderCh chan<- NetworkMessage) {
	for {
		select {
		case btn := <-drv_buttons:
			orderToLocalElevatorCh <- btn
		}
	}
}
