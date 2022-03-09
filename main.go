package main

//	"Driver-go/elevio"
import (
	"Driver-go/elevio"
	"fmt"
)

func main() {

	numFloors := 4
	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Down
	//elevio.SetMotorDirection(d)
	//go fsm.InitializeElevator()
	//go InitializeElevator()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	arrivedAtFloor := make(chan int)
	newOrderChan := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go ElevatorStateMachine(
		newOrderChan,
		arrivedAtFloor)

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
		}
	}

}

/* if init == 1 {
	d = elevio.MD_Stop
	init = 0
} else {
	if a == numFloors-1 {
		d = elevio.MD_Down
	} else if a == 0 {
		d = elevio.MD_Up
	}
}

elevio.SetMotorDirection(d) */

/*
//go onFloorArrival(a)

/* case a := <-drv_obstr:
fmt.Printf("%+v\n", a)
/*if a {
	elevio.SetMotorDirection(elevio.MD_Stop)
} else {
	elevio.SetMotorDirection(d)
}*/

/* case a := <-drv_stop:
fmt.Printf("%+v\n", a)
for f := 0; f < numFloors; f++ {
	for b := elevio.ButtonType(0); b < 3; b++ {
		elevio.SetButtonLamp(b, f, false)
	}
} */
