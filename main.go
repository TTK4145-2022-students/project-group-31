package main

//	"Driver-go/elevio"
import (
	"Driver-go/elevio"
	"Network-go/network/peers"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	//UNNECCESARY RAND
	rand.Seed(time.Now().UnixNano())
	numFloors := 4
	fmt.Println(os.Args)
	if len(os.Args) > 2 {
		port := os.Args[2]
		addr := "localhost:" + port
		elevio.Init(addr, numFloors)
	} else {
		elevio.Init("localhost:15657", numFloors)
	}

	drv_buttons := make(chan elevio.ButtonEvent, 1)
	drv_floors := make(chan int, 1)
	drv_obstr := make(chan bool, 1)
	drv_stop := make(chan bool, 1)

	newOrderChan := make(chan elevio.ButtonEvent, 1)
	networkOrder := make(chan elevio.ButtonEvent, 1)
	getElevChan := make(chan Elevator)
	elevatorUpdateChan := make(chan Elevator, 1)
	elevatorNetworkChan := make(chan ElevatorNetwork)

	elevatorInitializedChan := make(chan bool)

	redistributeOrdersChan := make(chan ElevatorNetwork, 1)
	localElevIDChan := make(chan string)
	distributeOrderChan := make(chan ElevatorMessage, 1)
	networkUpdateChan := make(chan NetworkMessage, 1)

	updateConnectionsChan := make(chan peers.PeerUpdate, 1)
	updateNewElevatorChan := make(chan ElevatorMessage, 1)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go ElevatorStateMachine(
		newOrderChan,
		drv_floors,
		drv_obstr,
		elevatorUpdateChan,
		getElevChan,
		elevatorInitializedChan)
	// Wait fo the elevator to be initialized before doing anything else
	<-elevatorInitializedChan

	go NetworkCommunication(
		elevatorUpdateChan,
		localElevIDChan,
		distributeOrderChan,
		networkUpdateChan,
		updateConnectionsChan,
		updateNewElevatorChan)

	go ElevatorNetworkStateMachine(
		localElevIDChan,
		elevatorNetworkChan,
		networkUpdateChan,
		networkOrder,
		updateConnectionsChan,
		getElevChan,
		redistributeOrdersChan,
		updateNewElevatorChan)

	go OrderDistributor(
		elevatorNetworkChan,
		drv_buttons,
		distributeOrderChan,
		newOrderChan,
		networkOrder,
		localElevIDChan,
		redistributeOrdersChan)

	select {
	//:(
	}
}
