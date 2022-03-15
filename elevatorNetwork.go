package main

import (
	"Driver-go/elevio"
	"fmt"
	"strconv"
)

const MAX_NUMBER_OF_ELEVATORS = 3

type ElevatorNetworkModule struct {
	Elevator  Elevator
	Connected bool
}

//Create struct for the ability to send network as parameter and on channel without writing as array
type ElevatorNetwork struct {
	ElevatorModules [MAX_NUMBER_OF_ELEVATORS]ElevatorNetworkModule
	//Connected          bool //Unsure wether to have or not. Connected to network for each elevator gives us the same info if EN[id].Connected=false
}

func ElevatorNetworkStateMachine(
	localElevIDChan <-chan string,
	elevatorNetworkChan chan<- ElevatorNetwork,
	updateElevatorChan chan<- Elevator,
	networkUpdateChan <-chan NetworkMessage,
	networkOrder chan<- elevio.ButtonEvent) {

	var elevatorNetwork ElevatorNetwork
	localElevID := <-localElevIDChan
	InitializeElevatorNetwork(&elevatorNetwork, localElevID)
	/* id, _ := strconv.Atoi(localElevID)
	elevatorNetwork.ElevatorModules[id].Connected = true
	elevatorNetwork.ElevatorModules[id].Elevator = <-updateElevatorChan */
	for {
		select {
		case elevatorNetworkChan <- elevatorNetwork:
			fmt.Println("Sendt EN")
		case ntwrkMsg := <-networkUpdateChan: //MAYBE SEND ENTIRE MESSAGE INSTEAD OF ONLY ELEV MESSAGE
			fmt.Println("Received elevMSG")
			elevID, _ := strconv.Atoi(ntwrkMsg.ElevatorMessage.ElevatorId)
			e := ntwrkMsg.ElevatorMessage.Elevator

			//Switch case may be moved outside if we want to do specific things for eac MT. Not necessary for most cases however because oftenmost update everything but orders
			if ntwrkMsg.ElevatorMessage.ElevatorId == localElevID {
				switch ntwrkMsg.MessageType {
				case MT_NewOrder:
					fmt.Println("New Order")
					//var newOrder elevio.ButtonEvent
					//Compare and find new order
					for f := 0; f < NUM_FLOORS; f++ {
						for btn := 0; btn < NUM_BUTTONS; btn++ {
							if e.Requests[f][btn] && !elevatorNetwork.ElevatorModules[elevID].Elevator.Requests[f][btn] {
								//Send order
								fmt.Printf("Found new order from network at: %#v", f)
								fmt.Printf("and button type at: %#v\n", elevio.ButtonType(btn))
								newOrder := elevio.ButtonEvent{f, elevio.ButtonType(btn)} //Why warning??
								networkOrder <- newOrder
							}
						}
					}
				}
				//Update ElevNetwork
				elevatorNetwork.ElevatorModules[elevID].Elevator = e
			}
		}

	}
}

func InitializeElevatorNetwork(en *ElevatorNetwork, localElevID string) {
	for i := 0; i < MAX_NUMBER_OF_ELEVATORS; i++ {
		InitializeElevator(&en.ElevatorModules[i].Elevator)
	}
	/* id, _ := strconv.Atoi(localElevID)
	en.ElevatorModules[id].Connected = true
	en.ElevatorModules[id] = <-updateElevatorChan */
}
