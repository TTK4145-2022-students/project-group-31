package main

import (
	"Driver-go/elevio"
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
	Orders    [NUM_FLOORS][NUM_BUTTONS]bool
	Behavior  ElevatorBehavior
	Online    bool
}

func (elevator *Elevator) Initialize() {
	elevator.Floor = 0
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Initialize
	elevator.SetAllLights()

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
func (e Elevator) SetAllLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, e.Orders[floor][btn])
		}
	}
}

//As we implement EN we want to only turn on cab lights and refuse to take hall
/* func (e Elevator) SetCabLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, e.Orders[floor][elevio.BT_Cab])
	}
} */

func ElevatorFSM(
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	orderToLocalElevatorCh <-chan elevio.ButtonEvent,
	initialElevator chan<- Elevator,
	elevatorStateChangeCh chan<- Elevator) {

	var elevator Elevator
	var obstructed bool
	elevator.Initialize()

	var doorClose <-chan time.Time

	var assumeMotorStop <-chan time.Time
	for {
		select {
		case order := <-orderToLocalElevatorCh:
			switch elevator.Behavior {
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
			}
			elevator.SetAllLights()

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
					elevator.SetAllLights()
				case EB_Moving:
					assumeMotorStop = time.After(TRAVEL_TIME * time.Second)
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(elevator.Direction)
				case EB_Idle:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(elevator.Direction)
				}
			}

		case newFloor := <-drv_floors:
			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)

			switch elevator.Behavior {
			case EB_Unavailable:
				if ShouldStop(elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearAtCurrentFloor(&elevator)
					elevator.SetAllLights()
					elevio.SetDoorOpenLamp(true)
					doorClose = time.After(DOOR_OPEN_DURATION * time.Second)
					elevator.Behavior = EB_DoorOpen
				} else {
					elevator.Behavior = EB_Moving
				}
			case EB_Moving:
				assumeMotorStop = time.After(TRAVEL_TIME * time.Second)
				if ShouldStop(elevator) {
					assumeMotorStop = nil
					elevio.SetMotorDirection(elevio.MD_Stop)
					clearAtCurrentFloor(&elevator)
					elevator.SetAllLights()

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

		case obstructed = <-drv_obstr:

		case <-assumeMotorStop:
			elevator.Behavior = EB_Unavailable
		}
	}
}
