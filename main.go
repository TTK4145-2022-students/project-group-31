package main

import (
	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"os"
)

func main() {

	//Initialize with id of the local elevator and a port. Use physical elevator if no simulator port is given
	localID := os.Args[1]
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
	numPeers := make(chan int, 1)

	networkMessageTx := make(chan NetworkMessage, 10)
	networkMessageRx := make(chan NetworkMessage, 10)

	elevatorInitialized := make(chan bool)
	elevatorStateChangeCh := make(chan Elevator)

	elevatorNetworkUpdateCh := make(chan [NUM_ELEVATORS]Elevator, 1)

	distributedOrderCh := make(chan NetworkMessage)
	addLocalOrder := make(chan elevio.ButtonEvent)

	reconnectedElevator := make(chan NetworkMessage, 10)
	updateElevatorNetworkCh := make(chan NetworkMessage, 10)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	go ElevatorFSM(
		drv_floors,
		drv_obstr,
		addLocalOrder,
		elevatorInitialized,
		elevatorStateChangeCh)
	//Waits for the Elevator to be initialized before starting the other go routines
	<-elevatorInitialized

	go bcast.Transmitter(TRANSCEIVER_PORT, networkMessageTx)
	go bcast.Receiver(TRANSCEIVER_PORT, networkMessageRx)

	go peers.Transmitter(PEERS_PORT, localID, peerTxEnable)
	go peers.Receiver(PEERS_PORT, peerUpdateCh)

	go orderDistributor(
		localID,
		drv_buttons,
		elevatorNetworkUpdateCh,
		addLocalOrder,
		distributedOrderCh)

	go NetworkTransceiver(
		localID,
		elevatorStateChangeCh,
		distributedOrderCh,
		reconnectedElevator,
		numPeers,
		networkMessageRx,
		updateElevatorNetworkCh,
		networkMessageTx)

	go ElevatorNetwork(
		localID,
		updateElevatorNetworkCh,
		peerUpdateCh,
		reconnectedElevator,
		elevatorNetworkUpdateCh,
		numPeers)

	select {}
}
