package main

import (
	"fmt"
	"time"
)

type MessageType int

const (
	MT_Acknowledge         MessageType = 0
	MT_ElevatorStateChange MessageType = 1
	MT_ReconnectedElevator MessageType = 2
	MT_NewOrder            MessageType = 3
)

type NetworkMessage struct {
	SenderID    string
	MessageType MessageType
	ElevatorID  string
	Elevator    Elevator
	Iter        int
}

func NetworkTransceiver(
	localID string,
	elevatorStateChangeCh <-chan Elevator,
	distributedOrderCh <-chan NetworkMessage,
	reconnectedElevator <-chan NetworkMessage,
	numPeers <-chan int,
	networkMessageRx <-chan NetworkMessage,
	updateElevatorNetworkCh chan<- NetworkMessage,
	networkMessageTx chan<- NetworkMessage) {

	lastTransmittedMessages := make(map[int]NetworkMessage)
	messageAcknowledgements := make(map[int]int)

	numOtherPeers := 0
	var resendMessages <-chan time.Time

	iteration := 0
	for {
		select {
		case msg := <-networkMessageRx:
			if msg.SenderID != localID {
				// Some way ignore acks if you do not want acks

				fmt.Println("Received message with type: ", msg.MessageType, "iteration: ", msg.Iter, "from ", msg.SenderID)
				if msg.MessageType == MT_Acknowledge {
					_, exists := messageAcknowledgements[msg.Iter]
					if exists {
						messageAcknowledgements[msg.Iter] = messageAcknowledgements[msg.Iter] + 1
						if messageAcknowledgements[msg.Iter] == numOtherPeers {
							fmt.Println("This message has been acknowledged")
							delete(lastTransmittedMessages, msg.Iter)
							delete(messageAcknowledgements, msg.Iter)
						}
					}
				} else {
					updateElevatorNetworkCh <- msg
					msg.MessageType = MT_Acknowledge
					msg.SenderID = localID
					networkMessageTx <- msg
				}
			}

		case elevator := <-elevatorStateChangeCh:
			iteration++
			msg := NetworkMessage{
				SenderID:    localID,
				MessageType: MT_ElevatorStateChange,
				ElevatorID:  localID,
				Elevator:    elevator,
				Iter:        iteration}
			updateElevatorNetworkCh <- msg

			if numOtherPeers > 0 {
				networkMessageTx <- msg
				lastTransmittedMessages[msg.Iter] = msg
				messageAcknowledgements[msg.Iter] = 0
				resendMessages = time.After(TIME_TO_RESEND * time.Millisecond)
			}

		case msg := <-distributedOrderCh:
			iteration++
			msg.Iter = iteration
			updateElevatorNetworkCh <- msg

			if numOtherPeers > 0 {
				networkMessageTx <- msg
				lastTransmittedMessages[msg.Iter] = msg
				messageAcknowledgements[msg.Iter] = 0
				resendMessages = time.After(TIME_TO_RESEND * time.Millisecond)
			}

		case msg := <-reconnectedElevator:
			iteration++
			msg.Iter = iteration
			updateElevatorNetworkCh <- msg

			if numOtherPeers > 0 {
				networkMessageTx <- msg
				lastTransmittedMessages[msg.Iter] = msg
				messageAcknowledgements[msg.Iter] = 0
				resendMessages = time.After(TIME_TO_RESEND * time.Millisecond)
			}

		case <-resendMessages:
			fmt.Println("Resend")
			if len(lastTransmittedMessages) > 0 {
				for iter, msg := range lastTransmittedMessages {
					messageAcknowledgements[iter] = 0
					networkMessageTx <- msg
					resendMessages = time.After(TIME_TO_RESEND * time.Millisecond)
					fmt.Println("Resendt message with type: ", msg.MessageType, "iteration: ", iter)
					break
				}
			} else {
				resendMessages = nil
			}

		case np := <-numPeers:
			numOtherPeers = np - 1
			fmt.Println("Num other peers: ", numOtherPeers)
		}
	}
}
