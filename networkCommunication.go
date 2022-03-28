package main

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"flag"
	"fmt"
	"time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type MessageType int

const (
	MT_Acknowledge    MessageType = 0
	MT_UpdateElevator MessageType = 1
	MT_NewElevator    MessageType = 2 // SHITTY NAME
	MT_NewOrder       MessageType = 3
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
	TimeStamp       time.Time
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

	networkMessageTx := make(chan NetworkMessage) //Might need buffer
	networkMessageRx := make(chan NetworkMessage) //Might need buffer
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(1412, networkMessageTx)
	go bcast.Receiver(1412, networkMessageRx)

	numPeers := 0

	unackedMessages := make(map[time.Time]int)
	lastSendtMessages := make(map[time.Time]NetworkMessage)

	var resendMessage <-chan time.Time

	wasStuck := false
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			fmt.Println("len:", len(p.Peers))
			numPeers = len(p.Peers)
			updateConnectionsChan <- p
		case rxMsg := <-networkMessageRx:
			/* fmt.Printf("Received: %#v\n", rxMsg) */
			fmt.Printf("Received network message\n")
			if rxMsg.MessageType == MT_Acknowledge {
				fmt.Println("Acknowledge")
				if rxMsg.SenderId != id {
					unackedMessages[rxMsg.TimeStamp] = unackedMessages[rxMsg.TimeStamp] + 1
				}
				if unackedMessages[rxMsg.TimeStamp] == numPeers-1 { //only care about the two others at a later date to minimilize packet loss oppurtunities
					delete(unackedMessages, rxMsg.TimeStamp)
					delete(lastSendtMessages, rxMsg.TimeStamp) // Deletes if it exists ignores if it doesn't
					networkUpdateChan <- rxMsg
				}

			} else if rxMsg.MessageType == MT_UpdateElevator {
				fmt.Println("UpdateElevator")
				networkUpdateChan <- rxMsg

			} else {
				unackedMessages[rxMsg.TimeStamp] = 0
				rxMsg.SenderId = id
				rxMsg.MessageType = MT_Acknowledge
				networkMessageTx <- rxMsg
			}

		case elevator := <-elevatorUpdateChan:
			txMsg := NetworkMessage{id, MT_UpdateElevator, ElevatorMessage{id, elevator}, time.Now()}

			if (elevator.Behavior == EB_MotorStop || elevator.Behavior == EB_DoorJam) && !wasStuck {
				peerTxEnable <- false
				wasStuck = true
			} else if (elevator.Behavior != EB_MotorStop && elevator.Behavior != EB_DoorJam) && wasStuck {
				peerTxEnable <- true
				wasStuck = false
				fmt.Println("We back bby!")
			}
			networkMessageTx <- txMsg

			fmt.Printf("Sendt elevator update\n")
		case elevatorMsg := <-distributeOrderChan:
			txMsg := NetworkMessage{id, MT_NewOrder, elevatorMsg, time.Now()}
			networkMessageTx <- txMsg

			lastSendtMessages[txMsg.TimeStamp] = txMsg
			resendMessage = time.After(TIME_TO_RESEND * time.Millisecond)
			fmt.Printf("Sendt order distribute\n")

		case elevatorMsg := <-updateNewElevatorChan:
			txMsg := NetworkMessage{id, MT_NewElevator, elevatorMsg, time.Now()}
			networkMessageTx <- txMsg
			lastSendtMessages[txMsg.TimeStamp] = txMsg
			resendMessage = time.After(TIME_TO_RESEND * time.Millisecond)
			fmt.Printf("Sendt new elevator update\n")

		case localElevIDChan <- id:
			fmt.Printf("Sendt id\n")
		case <-resendMessage:
			var messageToResend NetworkMessage
			var latestMessageTime time.Time
			for timeStamp, networkMsg := range lastSendtMessages {
				//Prioritize new elevator packetlosses over new order packetlosses
				if networkMsg.MessageType == MT_NewElevator {
					messageToResend = networkMsg
					break
				}
				if timeStamp.Before(latestMessageTime) {
					messageToResend = networkMsg
				}
			}
			networkMessageTx <- messageToResend
			fmt.Printf("Resend\n")
		}
	}
}
