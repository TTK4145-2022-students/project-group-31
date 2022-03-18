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
	networkUpdateChan <-chan ElevatorMessage,
	networkOrder chan<- elevio.ButtonEvent,
	updateConnectionsChan <-chan peers.PeerUpdate,
	updateElevatorChan chan<- Elevator,
	reconnectedElevChan <-chan Elevator,
	redistributeOrdersChan chan<- ElevatorNetwork) {

	var elevatorNetwork ElevatorNetwork
	localElevID := <-localElevIDChan
	elevatorNetwork.OnlyLocal = true //Default behavior
	//InitializeElevatorNetwork(&elevatorNetwork, localElevID)
	for {
		select {
		case elevatorNetworkChan <- elevatorNetwork:
			fmt.Println("Sendt EN")
		case ntwrkMsg := <-networkUpdateChan:
			fmt.Println("Received elevMSG")
			elevID, _ := strconv.Atoi(ntwrkMsg.ElevatorId)
			e := ntwrkMsg.Elevator
			if ntwrkMsg.ElevatorId == localElevID {
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
				updateElevatorChan <- ntwrkMsg.Elevator
			}
			//Update ElevNetwork
			elevatorNetwork.ElevatorModules[elevID].Elevator = e
			elevatorNetwork.SetAllHallLights(localElevID)
			PrintElevatorNetwork(elevatorNetwork)
		//Elevator becomes online or offline
		case p := <-updateConnectionsChan:
			if len(p.Peers) == 0 {
				elevatorNetwork.OnlyLocal = true
				fmt.Printf("ElevatorNetwork is now running in only local mode\n")
			} else {
				elevatorNetwork.OnlyLocal = false
				//id, _ := strconv.Atoi(p.New)
				//elevatorNetwork.ElevatorModules[id].Elevator = <-reconnectedElevChan
				fmt.Printf("ElevatorNetwork is now online\n")
			}
			if p.New != "" {
				if p.New == localElevID && len(p.Peers) != 1 {
					//Update myself and Update my perception of the others
					fmt.Println("Update myself from other nodes")
				}
				id, _ := strconv.Atoi(p.New)
				elevatorNetwork.ElevatorModules[id].Connected = true
				fmt.Printf("Elevator: %#v is now online\n", id)
				// If new
			}
			for _, lostElevID := range p.Lost {
				id, _ := strconv.Atoi(lostElevID)
				elevatorNetwork.ElevatorModules[id].Connected = false
				fmt.Printf("Elevator: %#v is now offline\n", id)
				redistributeOrdersChan <- elevatorNetwork
			}
			PrintElevatorNetwork(elevatorNetwork)
		}
	}
}

func InitializeElevatorNetwork(en *ElevatorNetwork, localElevID string) {
	for i := 0; i < MAX_NUMBER_OF_ELEVATORS; i++ {
		InitializeElevator(&en.ElevatorModules[i].Elevator)
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
