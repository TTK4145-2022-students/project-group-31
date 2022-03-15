package main

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	MT_Acknowledge     MessageType = 0
	MT_NewOrder        MessageType = 1
	MT_CompletedOrder  MessageType = 2 //These two could be just update elevator since we don't separate in the logic atm
	MT_ArrivedAtFloor  MessageType = 3 // maybe should be more actively used to later be able to clear orders
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
	distributeOrderChan <-chan ElevatorMessage,
	networkUpdateChan chan<- NetworkMessage) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
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

	peerCount := 0
	var lastReceivedMsg NetworkMessage
	var lastTransmittedMsg NetworkMessage
	var receivedAcks [MAX_NUMBER_OF_ELEVATORS]bool
	var transmitAgain <-chan time.Time
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
			peerCount = len(p.Peers)
		case a := <-networkMessageRx:
			//SIMULATE PACKET LOSS, remove if  when no longer testing
			if rand.Intn(10) == 1 {
				fmt.Println("PACKET LOST OH NO")
				break
			}
			fmt.Printf("Received: %#v\n", a)
			AckCount := 0
			if a.MessageType == MT_Acknowledge {
				intSenderId, _ := strconv.Atoi(a.SenderId)
				receivedAcks[intSenderId] = true
				for _, ack := range receivedAcks {
					if ack {
						AckCount += 1
					}
				}

				if AckCount == peerCount {
					networkUpdateChan <- lastReceivedMsg
					transmitAgain = nil
				}
			} else {
				lastReceivedMsg = a
				a.MessageType = MT_Acknowledge
				networkMessageTx <- a
				transmitAgain = time.After(100 * time.Millisecond)
			}

		case localElevIDChan <- id:
			fmt.Printf("Sendt id")
		case elevMsg := <-distributeOrderChan:
			networkMessage := NetworkMessage{id, MT_NewOrder, elevMsg}
			lastTransmittedMsg = networkMessage
			networkMessageTx <- networkMessage
			transmitAgain = time.After(100 * time.Millisecond)

		case elev := <-elevatorChan:
			networkMessage := NetworkMessage{id, MT_ArrivedAtFloor, ElevatorMessage{id, elev}}
			lastTransmittedMsg = networkMessage
			networkMessageTx <- networkMessage
			transmitAgain = time.After(100 * time.Millisecond)
		case a := <-transmitAgain:
			fmt.Printf("Did not receive all acks within 100 millisecond at: %#v\n", a)
			//Send the last message before sending ack again but ONLY the elevator that last sendt and not everybody
			//TEMPORARY send again
			networkMessageTx <- lastTransmittedMsg
		}

	}
}
