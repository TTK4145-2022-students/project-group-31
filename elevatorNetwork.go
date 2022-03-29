package main

import (
	"Network-go/network/peers"
	"fmt"
)

func ElevatorNetwork(
	id string,
	initialLocalElevator Elevator,
	updateElevatorNetworkCh <-chan NetworkMessage,
	peerUpdateCh <-chan peers.PeerUpdate,
	reconnectedElevator chan<- NetworkMessage,
	elevatorNetworkUpdateCh chan<- [NUM_ELEVATORS]Elevator) {

	var elevatorNetwork [NUM_ELEVATORS]Elevator
	fmt.Println(elevatorNetwork)
}
