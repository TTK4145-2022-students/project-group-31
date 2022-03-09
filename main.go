package main

//	"Driver-go/elevio"
import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4
	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	arrivedAtFloor := make(chan int)
	obstructionChan := make(chan bool)
	newOrderChan := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go ElevatorStateMachine(
		newOrderChan,
		arrivedAtFloor,
		obstructionChan)

	for {
		select {
		case btn := <-drv_buttons:
			fmt.Printf("%+v\n", btn)
			newOrderChan <- btn
			//elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			arrivedAtFloor <- a
		case a := <-drv_obstr:
			fmt.Printf("Obstruction", a)
			obstructionChan <- a
		}
	}
}
