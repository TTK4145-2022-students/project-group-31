package main

import (
	"Driver-go/elevio"
	"time"
)

func ElevatorFSM(
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	addLocalOrder <-chan elevio.ButtonEvent,
	elevatorInitialized chan<- bool,
	elevatorStateChangeCh chan<- Elevator) {

	var elevator Elevator
	var obstructed bool
	elevator.Initialize()

	var doorClose <-chan time.Time

	var assumeMotorStop <-chan time.Time

	for {
		select {
		case order := <-addLocalOrder:
			switch elevator.Behavior {

			case EB_Unavailable:
				elevator.AddOrder(order)

			case EB_DoorOpen:
				if ShouldClearImmediately(elevator, order.Floor, order.Button) {
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
				} else {
					elevator.AddOrder(order)
				}

			case EB_Moving:
				elevator.AddOrder(order)
				assumeMotorStop = time.After(TRAVEL_TIME * time.Second)

			case EB_Idle:
				elevator.AddOrder(order)
				elevator.Direction, elevator.Behavior = ChooseDirection(elevator)
				switch elevator.Behavior {
				case EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					elevator.clearAtCurrentFloor()
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)

				case EB_Moving:
					elevio.SetMotorDirection(elevator.Direction)
					assumeMotorStop = time.After(TRAVEL_TIME * time.Second)
				}
				elevatorStateChangeCh <- elevator
			}
			elevator.SetCabLights()

		case <-doorClose:
			if obstructed {
				doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
				elevator.Behavior = EB_Unavailable

			} else {
				elevator.Direction, elevator.Behavior = ChooseDirection(elevator)
				switch elevator.Behavior {
				case EB_DoorOpen:
					elevator.clearAtCurrentFloor()
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
					elevator.SetCabLights()

				case EB_Moving:
					assumeMotorStop = time.After(TRAVEL_TIME * time.Second)
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(elevator.Direction)

				case EB_Idle:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(elevator.Direction)
				}
			}
			elevatorStateChangeCh <- elevator

		case newFloor := <-drv_floors:
			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)
			switch elevator.Behavior {
			case EB_Unavailable:
				if ShouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.clearAtCurrentFloor()
					elevator.SetCabLights()
					elevio.SetDoorOpenLamp(true)
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
					elevator.Behavior = EB_DoorOpen

				} else {
					elevator.Direction, elevator.Behavior = ChooseDirection(elevator)
					elevio.SetMotorDirection(elevator.Direction)
				}

			case EB_Moving:
				assumeMotorStop = time.After(TRAVEL_TIME * time.Second)
				if ShouldStop(elevator) {
					assumeMotorStop = nil
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevator.clearAtCurrentFloor()
					elevator.SetCabLights()
					elevio.SetDoorOpenLamp(true)
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
					elevator.Behavior = EB_DoorOpen
				}

			case EB_Initialize:
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevator.Behavior = EB_Idle
				elevator.Direction = elevio.MD_Stop
				elevatorInitialized <- true
			}
			elevatorStateChangeCh <- elevator

		case obstructed = <-drv_obstr:

		case <-assumeMotorStop:
			elevator.Behavior = EB_Unavailable
			elevatorStateChangeCh <- elevator
		}
	}
}
