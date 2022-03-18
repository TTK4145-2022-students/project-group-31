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

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	newOrderChan := make(chan elevio.ButtonEvent)
	networkOrder := make(chan elevio.ButtonEvent)
	elevatorUpdateChan := make(chan NetworkMessage)
	elevatorNetworkChan := make(chan ElevatorNetwork)
	localElevIDChan := make(chan string)
	distributeOrderChan := make(chan ElevatorMessage)
	networkUpdateChan := make(chan NetworkMessage)

	updateConnectionsChan := make(chan peers.PeerUpdate)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go ElevatorStateMachine(
		newOrderChan,
		drv_floors,
		drv_obstr,
		elevatorUpdateChan)

	go Network(elevatorUpdateChan, localElevIDChan, distributeOrderChan, networkUpdateChan, updateConnectionsChan)
	go ElevatorNetworkStateMachine(localElevIDChan, elevatorNetworkChan, networkUpdateChan, networkOrder, updateConnectionsChan)

	go OrderDistributor(elevatorNetworkChan, drv_buttons, distributeOrderChan, newOrderChan, networkOrder, localElevIDChan)
	select {
	//:(
	}
}
