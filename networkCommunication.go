package main

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"flag"
	"fmt"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	MT_Acknowledge    MessageType = 0
	MT_UpdateElevator MessageType = 1
)

//Most likely an unneccessary struct and can just implement it in Network Message with ElevatorID
type ElevatorMessage struct {
	ElevatorId string
	Elevator   Elevator
}
type NetworkMessage struct {
	SenderId        string
	MessageType     MessageType
	ElevatorMessage ElevatorMessage
}

func NetworkCommunication(
	elevatorUpdateChan <-chan Elevator,
	localElevIDChan chan<- string,
	distributeOrderChan <-chan ElevatorMessage,
	networkUpdateChan chan<- NetworkMessage,
	updateConnectionsChan chan<- peers.PeerUpdate,
	updateNewElevatorChan <-chan ElevatorMessage) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run /. -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(2305, id, peerTxEnable)
	go peers.Receiver(2305, peerUpdateCh)

	networkMessageTx := make(chan NetworkMessage)
	networkMessageRx := make(chan NetworkMessage)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(1412, networkMessageTx)
	go bcast.Receiver(1412, networkMessageRx)

	/* peerCount := 0
	AckCount := 0
	var lastReceivedMsg NetworkMessage
	var lastTransmittedMsg NetworkMessage
	var receivedAcks [MAX_NUMBER_OF_ELEVATORS]bool
	var transmitAgain <-chan time.Time */
	//transmitAgain = time.After(1 * time.Second)

	//transmitAgain = nil
	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			fmt.Println("len:", len(p.Peers))
			//peerCount = len(p.Peers)
			updateConnectionsChan <- p
		case rxMsg := <-networkMessageRx:
			/* fmt.Printf("Received: %#v\n", rxMsg) */
			fmt.Printf("Received network message\n")
			networkUpdateChan <- rxMsg
		case elevator := <-elevatorUpdateChan:
			txMsg := NetworkMessage{id, MT_UpdateElevator, ElevatorMessage{id, elevator}}
			networkMessageTx <- txMsg
			/* fmt.Printf("Sendt: %#v\n", txMsg) */
			fmt.Printf("Sendt elevator update\n")
		case elevatorMsg := <-distributeOrderChan:
			txMsg := NetworkMessage{id, MT_UpdateElevator, elevatorMsg}
			networkMessageTx <- txMsg
			/* fmt.Printf("Sendt: %#v\n", txMsg) */
			fmt.Printf("Sendt order distribute\n")
		case elevatorMsg := <-updateNewElevatorChan:
			txMsg := NetworkMessage{id, MT_UpdateElevator, elevatorMsg}
			networkMessageTx <- txMsg
			/* fmt.Printf("Sendt: %#v\n", txMsg) */
			fmt.Printf("Sendt new elevator update\n")
		case localElevIDChan <- id:
			fmt.Printf("Sendt id\n")
		}
	}
}
