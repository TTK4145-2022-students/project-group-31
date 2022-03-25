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
	updateConnectionsChan <-chan peers.PeerUpdate,
	getElevChan <-chan Elevator,
	redistributeOrdersChan chan<- ElevatorNetwork,
	updateNewElevatorChan chan<- ElevatorMessage) {

	var elevatorNetwork ElevatorNetwork

	localElevID := <-localElevIDChan
	localElevIDInt, _ := strconv.Atoi(localElevID)

	fmt.Println("INIT NETWORK")
	elevatorNetwork.OnlyLocal = true //Default behavior
	for i := 0; i < MAX_NUMBER_OF_ELEVATORS; i++ {
		elevatorNetwork.ElevatorModules[i].Elevator.Behavior = EB_Initialize
	}
	elevatorNetwork.ElevatorModules[localElevIDInt].Elevator = <-getElevChan

	for {
		select {
		case elevatorNetworkChan <- elevatorNetwork:
			fmt.Println("Sendt EN")

		case ntwrkMsg := <-networkUpdateChan:
			fmt.Println("Received elevMSG")
			elevID, _ := strconv.Atoi(ntwrkMsg.ElevatorMessage.ElevatorId)
			e := ntwrkMsg.ElevatorMessage.Elevator
			if ntwrkMsg.ElevatorMessage.ElevatorId == localElevID {
				for f := 0; f < NUM_FLOORS; f++ {
					for btn := 0; btn < NUM_BUTTONS; btn++ {
						if e.Requests[f][btn] && !elevatorNetwork.ElevatorModules[elevID].Elevator.Requests[f][btn] {
							//Send order
							fmt.Printf("Found new order from network at: %#v", f)
							fmt.Printf("and button type at: %#v\n", elevio.ButtonType(btn))
							newOrder := elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(btn)}
							networkOrder <- newOrder
						}
					}
				}
				if ntwrkMsg.MessageType != MT_NewElevator {
					elevatorNetwork.ElevatorModules[elevID].Elevator = e
				} else {
					elevatorNetwork.ElevatorModules[elevID].Elevator.Requests = e.Requests
				}
			} else {
				//Overwrite ourself a the other elevator
				elevatorNetwork.ElevatorModules[elevID].Elevator = e
			}

			elevatorNetwork.SetAllHallLights(localElevID)
			PrintElevatorNetwork(elevatorNetwork)

		//Elevator becomes online or offline
		case p := <-updateConnectionsChan:

			//Only local mode
			if len(p.Peers) == 0 {
				elevatorNetwork.OnlyLocal = true
				fmt.Printf("ElevatorNetwork is now running in only local mode\n")
			} else {
				elevatorNetwork.OnlyLocal = false
				fmt.Printf("ElevatorNetwork is now online\n")
			}

			//Sending orders to each other on startup or reconnect
			//If you are not the new one send all uninit elevators
			if p.New != localElevID && p.New != "" {
				fmt.Println("Found new other elevator")
				for id := 0; id < MAX_NUMBER_OF_ELEVATORS; id++ {
					elevator := elevatorNetwork.ElevatorModules[id].Elevator
					if elevator.Behavior != EB_Initialize {
						updateNewElevatorChan <- ElevatorMessage{Elevator: elevator, ElevatorId: strconv.Itoa(id)}
						fmt.Println("Sendt new elevator: ", id)
					}
				}
				id, _ := strconv.Atoi(p.New)
				elevatorNetwork.ElevatorModules[id].Connected = true
			} /* else if p.New != "" { //IF you are the new one send only yourself if you have info
				fmt.Println("Found new my self elevator")
				elevator := elevatorNetwork.ElevatorModules[localElevIDInt].Elevator
				id, _ := strconv.Atoi(p.New)
				elevatorNetwork.ElevatorModules[id].Connected = true
				if elevator.Behavior != EB_Initialize {
					updateNewElevatorChan <- ElevatorMessage{Elevator: elevator, ElevatorId: localElevID}
					fmt.Println("Sendt new myself elevator: ")
				}
			} */
			//Losing elevators
			for _, lostElevID := range p.Lost {
				id, _ := strconv.Atoi(lostElevID)
				elevatorNetwork.ElevatorModules[id].Connected = false
				fmt.Printf("Elevator: %#v is now offline\n", id)
				redistributeOrdersChan <- elevatorNetwork
				//Maybe scary no confirmation that the orders have been successfully redistributed. Maybe not critical
				for floor := 0; floor < NUM_FLOORS; floor++ {
					for btn := elevio.ButtonType(0); btn < NUM_BUTTONS-1; btn++ {
						elevatorNetwork.ElevatorModules[id].Elevator.Requests[floor][btn] = false
					}
				}
			}
			PrintElevatorNetwork(elevatorNetwork)
		}
	}
}

func (en ElevatorNetwork) SetAllHallLights(localElevID string) {
	//Create lights matrix for hall calls. Cab calls are turned on and off locally
	var lights [NUM_FLOORS][NUM_BUTTONS - 1]bool
	for id := 0; id < MAX_NUMBER_OF_ELEVATORS; id++ {
		elevator := en.ElevatorModules[id].Elevator
		for floor := 0; floor < NUM_FLOORS; floor++ {
			for btn := elevio.ButtonType(0); btn < NUM_BUTTONS-1; btn++ {
				lights[floor][btn] = lights[floor][btn] || elevator.Requests[floor][btn]
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
func PrintElevatorNetwork(en ElevatorNetwork) {

	for id := 0; id < MAX_NUMBER_OF_ELEVATORS; id++ {
		fmt.Println("Elevator: ", id)
		elevator := en.ElevatorModules[id].Elevator
		fmt.Printf("| B: %+v", elevator.Behavior)
		fmt.Printf(" | D: %+v", elevator.Direction)
		fmt.Printf(" | F: %+v |\n", elevator.Floor)
		fmt.Println("   UP       DOWN     CAB")
		for floor := 0; floor < NUM_FLOORS; floor++ {
			fmt.Printf("|")
			for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
				fmt.Printf("  %+v  ", elevator.Requests[floor][btn])
			}
			fmt.Printf("|\n")
		}
	}
}

/* func IsInitializedElevator(elevator Elevator) bool {
	var initElevator Elevator
	InitializeElevator(&initElevator)

	allOrdersEqual := true
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
			if initElevator.Requests[floor][btn] != elevator.Requests[floor][btn] {
				allOrdersEqual = false
			}
		}
	}
	if elevator.Behavior == initElevator.Behavior &&
		elevator.Direction == initElevator.Direction &&
		elevator.Floor == initElevator.Floor &&
		allOrdersEqual {
		return true
	} else {
		return false
	}
} */
