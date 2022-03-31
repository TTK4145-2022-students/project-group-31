package main

import (
	"Driver-go/elevio"
	"Network-go/network/peers"
	"fmt"
	"strconv"
)

func ElevatorNetwork(
	localID string,
	initialLocalElevator Elevator,
	updateElevatorNetworkCh <-chan NetworkMessage,
	peerUpdateCh <-chan peers.PeerUpdate,
	reconnectedElevator chan<- NetworkMessage,
	elevatorNetworkUpdateCh chan<- [NUM_ELEVATORS]Elevator,
	numPeers chan<- int) {

	var elevatorNetwork [NUM_ELEVATORS]Elevator
	for id := 0; id < NUM_ELEVATORS; id++ {
		elevatorNetwork[id].Behavior = EB_Unavailable
	}

	var isLost [NUM_ELEVATORS]bool
	for {
		select {
		case networkMsg := <-updateElevatorNetworkCh:

			elevator := networkMsg.Elevator
			elevatorID, _ := strconv.Atoi(networkMsg.ElevatorID)

			if networkMsg.SenderID == networkMsg.ElevatorID && !isLost[elevatorID] {
				elevatorNetwork[elevatorID] = elevator

			} else {
				for id := 0; id < NUM_ELEVATORS; id++ {
					for floor := 0; floor < NUM_FLOORS; floor++ {
						for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
							if id == elevatorID {
								elevatorNetwork[id].Orders[floor][btn] = elevatorNetwork[id].Orders[floor][btn] || elevator.Orders[floor][btn]
							}
						}
					}
				}
			}
			fmt.Println("Received Network Message of type", networkMsg.MessageType)
			fmt.Println("From: ", networkMsg.SenderID, "About: ", networkMsg.ElevatorID)
			PrintElevatorNetwork(elevatorNetwork)
			elevatorNetworkUpdateCh <- elevatorNetwork
			SetAllHallLights(elevatorNetwork)

		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			numPeers <- len(p.Peers)
			if p.New != "" && p.New != localID {

				fmt.Println("Sending elevator ", localID, " myself")
				localIDInt, _ := strconv.Atoi(localID)
				elevatorNetwork[localIDInt].Print()
				reconnectedElevator <- NetworkMessage{
					SenderID:    localID,
					MessageType: MT_ReconnectedElevator,
					ElevatorID:  localID,
					Elevator:    elevatorNetwork[localIDInt]}

				fmt.Println("Sending new elevator ", p.New, " info")
				id, _ := strconv.Atoi(p.New)
				elevatorNetwork[id].Print()
				reconnectedElevator <- NetworkMessage{
					SenderID:    localID,
					MessageType: MT_ReconnectedElevator,
					ElevatorID:  p.New,
					Elevator:    elevatorNetwork[id]}

				isLost[id] = false
			}

			if len(p.Lost) > 0 {
				for _, lost := range p.Lost {
					id, _ := strconv.Atoi(lost)
					elevatorNetwork[id].Behavior = EB_Unavailable
					isLost[id] = true
				}
				elevatorNetworkUpdateCh <- elevatorNetwork
			}

			if len(p.Peers) == 0 {
				localIDInt, _ := strconv.Atoi(localID)
				isLost[localIDInt] = false
			}
		}
	}
}

func PrintElevatorNetwork(elevatorNetwork [NUM_ELEVATORS]Elevator) {
	fmt.Printf("|")
	for id := 0; id < NUM_ELEVATORS; id++ {

		elevator := elevatorNetwork[id]
		fmt.Printf(" B:%+v ", elevator.Behavior)

		fmt.Printf(" D:%+v ", elevator.Direction)

		fmt.Printf(" F:%+v ", elevator.Floor)
		if id == 0 {
			fmt.Printf("|")
		}
		if id == 1 {
			fmt.Printf(" |")
		}
		if id == 2 {
			fmt.Printf(" |")
		}
	}
	fmt.Println()
	fmt.Printf("| ")
	for id := 0; id < NUM_ELEVATORS; id++ {
		fmt.Printf(" UP  DOWN  CAB")
		if id == 0 {
			fmt.Printf("| ")
		}
		if id == 1 {
			fmt.Printf(" |")
		}
		if id == 2 {
			fmt.Printf("  |")
		}

	}
	fmt.Println()
	for floor := 0; floor < NUM_FLOORS; floor++ {
		fmt.Printf("|")
		for id := 0; id < NUM_ELEVATORS; id++ {
			elevator := elevatorNetwork[id]
			for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
				if elevator.Orders[floor][btn] {
					fmt.Printf("  %+v  ", 1)
				} else {
					fmt.Printf("  %+v  ", 0)
				}
			}
			fmt.Printf("| ")
		}
		fmt.Printf("\n")
	}
	fmt.Println()
}

func SetAllHallLights(elevatorNetwork [NUM_ELEVATORS]Elevator) {
	//Create lights matrix for hall calls. Cab calls are turned on and off locally
	var lights [NUM_FLOORS][NUM_BUTTONS - 1]bool
	for id := 0; id < NUM_ELEVATORS; id++ {
		elevator := elevatorNetwork[id]
		for floor := 0; floor < NUM_FLOORS; floor++ {
			for btn := elevio.ButtonType(0); btn < NUM_BUTTONS-1; btn++ {
				if elevator.Behavior != EB_Unavailable {
					lights[floor][btn] = lights[floor][btn] || elevator.Orders[floor][btn]
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
}
