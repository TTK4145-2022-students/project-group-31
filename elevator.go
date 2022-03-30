package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

type ElevatorBehavior int

const (
	EB_Idle        ElevatorBehavior = 0
	EB_DoorOpen    ElevatorBehavior = 1
	EB_Moving      ElevatorBehavior = 2
	EB_Unavailable ElevatorBehavior = 3
	EB_Initialize  ElevatorBehavior = 4
)

type Elevator struct {
	Floor     int
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior

	Orders [NUM_FLOORS][NUM_BUTTONS]bool

	Online bool
}

func (elevator *Elevator) Initialize() {
	elevator.Floor = 0
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Initialize
	elevator.SetCabLights()

	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)
}

func (e *Elevator) AddOrder(order elevio.ButtonEvent) {
	e.Orders[order.Floor][order.Button] = true
}
func (e *Elevator) RemoveOrder(order elevio.ButtonEvent) {
	e.Orders[order.Floor][order.Button] = false
}

//As we implement EN we want to only turn on cab lights and refuse to take hall
/* func (e Elevator) SetAllLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, e.Orders[floor][btn])
		}
	}
} */

//As we implement EN we want to only turn on cab lights and refuse to take hall
func (e Elevator) SetCabLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, e.Orders[floor][elevio.BT_Cab])
	}
}

func (elevator Elevator) Print() {
	fmt.Printf("| B: %+v", elevator.Behavior)
	fmt.Printf(" | D: %+v", elevator.Direction)
	fmt.Printf(" | F: %+v |\n", elevator.Floor)
	fmt.Println("   UP       DOWN     CAB")
	for floor := 0; floor < NUM_FLOORS; floor++ {
		fmt.Printf("|")
		for btn := elevio.ButtonType(0); btn < NUM_BUTTONS; btn++ {
			fmt.Printf("  %+v  ", elevator.Orders[floor][btn])
		}
		fmt.Printf("|\n")
	}
}

func ElevatorFSM(
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	addLocalOrder <-chan elevio.ButtonEvent,
	initialElevator chan<- Elevator,
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
					clearAtCurrentFloor(&elevator)
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
				case EB_Moving:
					elevio.SetMotorDirection(elevator.Direction)
					assumeMotorStop = time.After(TRAVEL_TIME * time.Second)
				}

				elevatorStateChangeCh <- elevator // Maybe send anyways
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
					clearAtCurrentFloor(&elevator)
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
					clearAtCurrentFloor(&elevator)
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
					clearAtCurrentFloor(&elevator)
					elevator.SetCabLights()

					elevio.SetDoorOpenLamp(true)
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
					elevator.Behavior = EB_DoorOpen
				}
			case EB_Initialize:
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevator.Behavior = EB_Idle
				elevator.Direction = elevio.MD_Stop
				initialElevator <- elevator
			}
			elevatorStateChangeCh <- elevator

		case obstructed = <-drv_obstr:

		case <-assumeMotorStop:
			elevator.Behavior = EB_Unavailable
			elevatorStateChangeCh <- elevator
		}
	}
}
