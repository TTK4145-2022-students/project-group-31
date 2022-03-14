package main

import (
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
	networkUpdateChan <-chan ElevatorMessage) {

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
		case elevMsg := <-networkUpdateChan: //MAYBE SEND ENTIRE MESSAGE INSTEAD OF ONLY ELEV MESSAGE
			fmt.Println("Received elevMSG")
			elevID, _ := strconv.Atoi(elevMsg.ElevatorId)
			elevatorNetwork.ElevatorModules[elevID].Elevator = elevMsg.Elevator
			/* if elevMsg.ElevatorId == localElevID {
				updateElevatorChan <- elevMsg.Elevator
			} */
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
