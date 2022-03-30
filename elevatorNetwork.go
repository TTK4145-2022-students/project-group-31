package main

import (
	"Driver-go/elevio"
	"Network-go/network/peers"
	"fmt"
	"strconv"
)

func ElevatorNetwork(
	localID string,
	initialLocalElevator Elevator,
	updateElevatorNetworkCh <-chan NetworkMessage,
	peerUpdateCh <-chan peers.PeerUpdate,
	reconnectedElevator chan<- NetworkMessage,
	elevatorNetworkUpdateCh chan<- [NUM_ELEVATORS]Elevator) {

	var elevatorNetwork [NUM_ELEVATORS]Elevator

	for {
		select {
		case networkMsg := <-updateElevatorNetworkCh:
			elevator := networkMsg.Elevator
			elevatorID, _ := strconv.Atoi(networkMsg.ElevatorID)
			if networkMsg.SenderID == networkMsg.ElevatorID {
				elevatorNetwork[elevatorID] = elevator
			} else {
				for id := 0; id < NUM_ELEVATORS; id++ {
					for floor := 0; floor < NUM_FLOORS; floor++ {
						for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
							elevatorNetwork[id].Orders[floor][btn] = elevatorNetwork[id].Orders[floor][btn] || elevator.Orders[floor][btn]
						}
					}
				}
			}
			PrintElevatorNetwork(elevatorNetwork)
			//Redistribute orders and find new orders to local
		}
	}
}

func PrintElevatorNetwork(elevatorNetwork [NUM_ELEVATORS]Elevator) {
	fmt.Printf("|")
	for id := 0; id < NUM_ELEVATORS; id++ {

		elevator := elevatorNetwork[id]
		fmt.Printf(" B:%+v ", elevator.Behavior)

		fmt.Printf(" D:%+v ", elevator.Direction)

		fmt.Printf(" F:%+v ", elevator.Floor)
		if id == 0 {
			fmt.Printf("|")
		}
		if id == 1 {
			fmt.Printf(" |")
		}
		if id == 2 {
			fmt.Printf(" |")
		}
	}
	fmt.Println()
	fmt.Printf("| ")
	for id := 0; id < NUM_ELEVATORS; id++ {
		fmt.Printf(" UP  DOWN  CAB")
		if id == 0 {
			fmt.Printf("| ")
		}
		if id == 1 {
			fmt.Printf(" |")
		}
		if id == 2 {
			fmt.Printf("  |")
		}

	}
	fmt.Println()
	for floor := 0; floor < NUM_FLOORS; floor++ {
		fmt.Printf("|")
		for id := 0; id < NUM_ELEVATORS; id++ {
			elevator := elevatorNetwork[id]
			for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
				if elevator.Orders[floor][btn] {
					fmt.Printf("  %+v  ", 1)
				} else {
					fmt.Printf("  %+v  ", 0)
				}
			}
			fmt.Printf("| ")
		}
		fmt.Printf("\n")
	}
	fmt.Println()
}
