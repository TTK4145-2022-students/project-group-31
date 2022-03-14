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
	MT_Acknowledge     MessageType = 0
	MT_NewOrder        MessageType = 1
	MT_CompletedOrder  MessageType = 2
	MT_ArrivedAtFloor  MessageType = 3
	MT_InitialElevator MessageType = 4
)

type ElevatorMessage struct {
	ElevatorId string
	Elevator   Elevator
}
type NetworkMessage struct {
	SenderId        string
	MessageType     MessageType
	ElevatorMessage ElevatorMessage
}

func Network(
	elevatorChan <-chan Elevator,
	localElevIDChan chan<- string,
	elevatorMessageChan <-chan ElevatorMessage,
	networkUpdateChan chan<- ElevatorMessage) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	/*if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}*/

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(2305, id, peerTxEnable)
	go peers.Receiver(2305, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	//helloTx := make(chan HelloMsg)
	//helloRx := make(chan HelloMsg)

	networkMessageTx := make(chan NetworkMessage)
	networkMessageRx := make(chan NetworkMessage)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(1412, networkMessageTx)
	go bcast.Receiver(1412, networkMessageRx)

	// The example message. We just send one of these every second.
	/* go func() {

		for {
			currentElevator := <-elevatorChan
			networkMessage := NetworkMessage{id, MT_NewOrder, currentElevator}
			networkMessageTx <- networkMessage
			time.Sleep(5 * time.Second)
		}
	}() */

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-networkMessageRx:
			fmt.Printf("Received: %#v\n", a)
			switch a.MessageType {
			case MT_ArrivedAtFloor:
				networkUpdateChan <- a.ElevatorMessage
			case MT_NewOrder:
				networkUpdateChan <- a.ElevatorMessage
			}
		case localElevIDChan <- id:
			fmt.Printf("Sendt id")
		case elevMsg := <-elevatorMessageChan:
			networkMessage := NetworkMessage{id, MT_NewOrder, elevMsg}
			networkMessageTx <- networkMessage

		case elev := <-elevatorChan:
			networkMessage := NetworkMessage{id, MT_ArrivedAtFloor, ElevatorMessage{id, elev}}
			networkMessageTx <- networkMessage
		}
	}
}
