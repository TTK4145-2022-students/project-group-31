package main

import (
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
	TimeStamp   time.Time
}

func NetworkTransceiver(
	localID string,
	elevatorStateChangeCh <-chan Elevator,
	distributedOrderCh <-chan NetworkMessage,
	reconnectedElevator <-chan NetworkMessage,
	networkMessageRx <-chan NetworkMessage,
	updateElevatorNetworkCh chan<- NetworkMessage,
	networkMessageTx chan<- NetworkMessage) {

	lastTransmittedMessages := make(map[MessageType]NetworkMessage)

	var resendMessages <-chan time.Time
	//resendMessages = time.After(TIME_TO_RESEND * time.Millisecond)
	for {
		select {
		case msg := <-networkMessageRx:
			updateElevatorNetworkCh <- msg
		case elevator := <-elevatorStateChangeCh:
			networkMsg := NetworkMessage{SenderID: localID, MessageType: MT_ElevatorStateChange, ElevatorID: localID, Elevator: elevator, TimeStamp: time.Now()}
			lastTransmittedMessages[MT_ElevatorStateChange] = networkMsg
			networkMessageTx <- networkMsg

		case msg := <-distributedOrderCh:
			lastTransmittedMessages[msg.MessageType] = msg
			networkMessageTx <- msg

		case msg := <-reconnectedElevator:
			lastTransmittedMessages[msg.MessageType] = msg
			networkMessageTx <- msg
		case <-resendMessages:
			/* var mostRecentTime time.Time
			var msgToSend NetworkMessage

			for _, msg := range lastTransmittedMessages {
				//ONLY SEND MOST RECENT
				if msg.TimeStamp.After(mostRecentTime) {
					mostRecentTime = msg.TimeStamp
					msgToSend = msg
				}
			}
			networkMessageTx <- msgToSend */
			for _, msg := range lastTransmittedMessages {
				networkMessageTx <- msg
			}

			resendMessages = time.After(TIME_TO_RESEND * time.Millisecond)
		}
	}
}
