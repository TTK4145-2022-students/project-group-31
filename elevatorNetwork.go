package main

import (
	"Driver-go/elevio"
	"Network-go/network/peers"
	"strconv"
)

func ElevatorNetwork(
	localID string,
	updateElevatorNetworkCh <-chan NetworkMessage,
	peerUpdateCh <-chan peers.PeerUpdate,
	reconnectedElevator chan<- NetworkMessage,
	elevatorNetworkChangeCh chan<- [NUM_ELEVATORS]Elevator,
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
				for floor := 0; floor < NUM_FLOORS; floor++ {
					for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
						elevatorNetwork[elevatorID].Orders[floor][btn] =
							elevatorNetwork[elevatorID].Orders[floor][btn] || elevator.Orders[floor][btn]
					}
				}
			}
			elevatorNetworkChangeCh <- elevatorNetwork
			SetAllHallLights(elevatorNetwork)

		case p := <-peerUpdateCh:
			numPeers <- len(p.Peers)
			if p.New != "" && p.New != localID {
				localIDInt, _ := strconv.Atoi(localID)
				reconnectedElevator <- NetworkMessage{
					SenderID:    localID,
					MessageType: MT_ReconnectedElevator,
					ElevatorID:  localID,
					Elevator:    elevatorNetwork[localIDInt]}

				id, _ := strconv.Atoi(p.New)
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
				elevatorNetworkChangeCh <- elevatorNetwork
			}

			if len(p.Peers) == 0 {
				localIDInt, _ := strconv.Atoi(localID)
				isLost[localIDInt] = false
			}
		}
	}
}

func SetAllHallLights(elevatorNetwork [NUM_ELEVATORS]Elevator) {
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

	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < NUM_BUTTONS-1; btn++ {
			elevio.SetButtonLamp(btn, floor, lights[floor][btn])
		}
	}
}
