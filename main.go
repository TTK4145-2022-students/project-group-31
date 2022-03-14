package main

//	"Driver-go/elevio"
import (
	"Driver-go/elevio"
)

func main() {

	numFloors := 4
	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	newOrderChan := make(chan elevio.ButtonEvent)
	elevatorChan := make(chan Elevator)
	elevatorNetworkChan := make(chan ElevatorNetwork)
	localElevIDChan := make(chan string)
	elevatorMessageChan := make(chan ElevatorMessage)
	networkUpdateChan := make(chan ElevatorMessage)
	updateElevatorChan := make(chan Elevator)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go ElevatorStateMachine(
		newOrderChan,
		drv_floors,
		drv_obstr,
		elevatorChan,
		updateElevatorChan)

	go Network(elevatorChan, localElevIDChan, elevatorMessageChan, networkUpdateChan)
	go ElevatorNetworkStateMachine(localElevIDChan, elevatorNetworkChan, updateElevatorChan, networkUpdateChan)

	go OrderDistributor(elevatorNetworkChan, drv_buttons, elevatorMessageChan, newOrderChan)
	for {
		//:(
	}
}
