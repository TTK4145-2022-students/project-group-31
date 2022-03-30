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
	for {
		select {
		case msg := <-networkMessageRx:
			//fmt.Println("Elevator ID: ", localID)
			//msg.Elevator.Print()
			updateElevatorNetworkCh <- msg
		case elevator := <-elevatorStateChangeCh:
			networkMsg := NetworkMessage{SenderID: localID, MessageType: MT_ElevatorStateChange, ElevatorID: localID, Elevator: elevator, TimeStamp: time.Now()}
			networkMessageTx <- networkMsg
		case msg := <-distributedOrderCh:
			networkMessageTx <- msg
		case elevatorMsg := <-reconnectedElevator:
			fmt.Println(elevatorMsg)
		}
	}
}
