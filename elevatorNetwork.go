package main

import (
	"Driver-go/elevio"
	"Network-go/network/peers"
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
	OnlyLocal       bool //Unsure wether to have or not. Connected to network for each elevator gives us the same info if EN[id].Connected=false
}

func ElevatorNetworkStateMachine(
	localElevIDChan <-chan string,
	elevatorNetworkChan chan<- ElevatorNetwork,
	networkUpdateChan <-chan NetworkMessage,
	networkOrder chan<- elevio.ButtonEvent,
	updateConnectionsChan <-chan peers.PeerUpdate) {

	var elevatorNetwork ElevatorNetwork
	localElevID := <-localElevIDChan
	//InitializeElevatorNetwork(&elevatorNetwork, localElevID)
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
					/* case MT_CompletedOrder:
					fmt.Println("Completed Order")
					//var newOrder elevio.ButtonEvent
					//Compare and find new order
					for f := 0; f < NUM_FLOORS; f++ {
						for btn := 0; btn < NUM_BUTTONS; btn++ {
							if !e.Requests[f][btn] && elevatorNetwork.ElevatorModules[elevID].Elevator.Requests[f][btn] {
								//Send order
								fmt.Printf("Found completed order from network at: %#v", f)
								fmt.Printf("and button type at: %#v\n", elevio.ButtonType(btn))
							}
						}
					} */
				}
			}
			//Update ElevNetwork
			elevatorNetwork.ElevatorModules[elevID].Elevator = e
			elevatorNetwork.SetAllLights(localElevID)
		//Elevator becomes online or offline
		case p := <-updateConnectionsChan:
			/* if len(p.Peers)==1{
				elevatorNetwork.OnlyLocal = false
				fmt.Printf("ElevatorNetwork is now running in only local mode\n")
			} */
			if p.New != "" {
				id, _ := strconv.Atoi(p.New)
				elevatorNetwork.ElevatorModules[id].Connected = true
				fmt.Printf("Elevator: %#v is now online\n", id)
				if p.New != localElevID {
					elevatorNetwork.OnlyLocal = true
					fmt.Printf("ElevatorNetwork is now online\n")
				}
				// If new 
			}
			for _, lostElevID := range p.Lost {
				id, _ := strconv.Atoi(lostElevID)
				elevatorNetwork.ElevatorModules[id].Connected = false
				fmt.Printf("Elevator: %#v is now offline\n", id)
				//LOGIC FOR RESTRIBUTING ORDERS
			}
		}
	}
}

func InitializeElevatorNetwork(en *ElevatorNetwork, localElevID string) {
	for i := 0; i < MAX_NUMBER_OF_ELEVATORS; i++ {
		InitializeElevator(&en.ElevatorModules[i].Elevator)
	}
}

func (en ElevatorNetwork) SetAllLights(localElevID string) {
	//Create lights matrix
	var lights [NUM_FLOORS][NUM_BUTTONS]bool
	for id := 0; id < MAX_NUMBER_OF_ELEVATORS; id++ {
		elevator := en.ElevatorModules[id].Elevator
		for floor := 0; floor < NUM_FLOORS; floor++ {
			for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
				/* if elevator.Requests[floor][btn] {
					lights[floor][btn] = true
				} else if elevator.Requests[floor][NUM_BUTTONS-1] {
					elevio.SetButtonLamp(NUM_BUTTONS-1, floor, elevator.Requests[floor][NUM_BUTTONS-1])
				} */
				lights[floor][btn] = lights[floor][btn] || elevator.Requests[floor][btn]
				if strconv.Itoa(id) == localElevID && btn == elevio.BT_Cab {
					elevio.SetButtonLamp(btn, floor, lights[floor][btn])
				}
			}
		}
	}
	//Turn on or off lights accordingly
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < NUM_BUTTONS-1; btn++ {
			elevio.SetButtonLamp(btn, floor, lights[floor][btn])
		}
	}

	/* for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
			hallOrder := false
			for id := 0; id < MAX_NUMBER_OF_ELEVATORS; id++ {
				if btn == elevio.BT_Cab && strconv.Itoa(id) == localElevID {
					elevio.SetButtonLamp(btn, floor, en.ElevatorModules[id].Elevator.Requests[floor][btn])
				} else if en.ElevatorModules[id].Elevator.Requests[floor][btn] {
					hallOrder = true
				}
			}
			elevio.SetButtonLamp(btn, floor, hallOrder)
		}
	} */
}
