package main

import (
	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"os"
)

func main() {
	//Initialize with id of the local elevator and a port. Use physical elevator if no simulator port is given
	id := os.Args[1]
	if len(os.Args) > 2 {
		port := os.Args[2]
		addr := "localhost:" + port
		elevio.Init(addr, NUM_FLOORS)
	} else {
		elevio.Init("localhost:15657", NUM_FLOORS)
	}

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	networkMessageTx := make(chan NetworkMessage)
	networkMessageRx := make(chan NetworkMessage)

	initialElevator := make(chan Elevator)
	elevatorStateChangeCh := make(chan Elevator)

	elevatorNetworkUpdateCh := make(chan [NUM_ELEVATORS]Elevator)

	distributedOrderCh := make(chan NetworkMessage)
	orderToLocalElevatorCh := make(chan elevio.ButtonEvent)

	reconnectedElevator := make(chan NetworkMessage)
	updateElevatorNetworkCh := make(chan NetworkMessage)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	go ElevatorFSM(
		drv_floors,
		drv_obstr,
		orderToLocalElevatorCh,
		initialElevator,
		elevatorStateChangeCh)
	//Waits for the Elevator to be initialized before starting the other go routines
	initialLocalElevator := <-initialElevator

	go peers.Transmitter(PEERS_PORT, id, peerTxEnable)
	go peers.Receiver(PEERS_PORT, peerUpdateCh)

	go bcast.Transmitter(TRANSCEIVER_PORT, networkMessageTx)
	go bcast.Receiver(TRANSCEIVER_PORT, networkMessageRx)

	go orderDistributor(
		id,
		drv_buttons,
		elevatorNetworkUpdateCh,
		orderToLocalElevatorCh,
		distributedOrderCh)

	go NetworkTransceiver(
		id,
		elevatorStateChangeCh,
		distributedOrderCh,
		reconnectedElevator,
		networkMessageRx,
		updateElevatorNetworkCh,
		networkMessageTx)
	go ElevatorNetwork(
		id,
		initialLocalElevator,
		updateElevatorNetworkCh,
		peerUpdateCh,
		reconnectedElevator,
		elevatorNetworkUpdateCh)

	select {}
}
