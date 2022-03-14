package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

const NUM_FLOORS = 4
const NUM_BUTTONS = 3
const DOOR_OPEN_DURATION = 3

type ElevatorBehavior int

const (
	EB_Idle     ElevatorBehavior = 0
	EB_DoorOpen ElevatorBehavior = 1
	EB_Moving   ElevatorBehavior = 2
)

type Elevator struct {
	Floor     int
	Direction elevio.MotorDirection
	Requests  [NUM_FLOORS][NUM_BUTTONS]bool
	Behavior  ElevatorBehavior
}

func (e Elevator) SetAllLights() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, e.Requests[floor][btn])
		}
	}
}

func (e *Elevator) AddOrder(btnFloor int, btnType elevio.ButtonType) {
	e.Requests[btnFloor][btnType] = true
}
func (e *Elevator) RemoveOrder(btnFloor int, btnType elevio.ButtonType) {
	e.Requests[btnFloor][btnType] = false
}

func ElevatorStateMachine(
	newOrderChan <-chan elevio.ButtonEvent,
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	elevatorChan chan<- Elevator,
	updateElevatorChan <-chan Elevator) {
	var elevator Elevator
	obstructed := false
	//timerFinishedChannel := make(chan int)
	InitializeElevator(&elevator)

	var doorClose <-chan time.Time

	//doorClose = nil

	for {
		select {

		case btn := <-newOrderChan:
			fmt.Printf("COST: %+v\n", CalculateCost(elevator, btn))
			btnFloor := btn.Floor
			btnType := btn.Button
			switch elevator.Behavior {
			case EB_DoorOpen:
				if ShouldClearImmediately(elevator, btnFloor, btnType) {
					//START TIMER
					doorClose = time.After(3 * time.Second)
				} else {
					elevator.AddOrder(btnFloor, btnType)
				}
			case EB_Moving:
				elevator.AddOrder(btnFloor, btnType)
			case EB_Idle:
				elevator.AddOrder(btnFloor, btnType)
				elevator.Direction, elevator.Behavior = NextAction(elevator)
				fmt.Printf("new order action; \n")
				fmt.Printf("EB: %+v\n", elevator.Behavior)
				fmt.Printf("DIR: %+v\n", elevator.Direction)
				switch elevator.Behavior {
				case EB_DoorOpen:
					//OPEN DOOR AND START TIMER
					elevio.SetDoorOpenLamp(true)
					clearAtCurrentFloor(&elevator)
					doorClose = time.After(3 * time.Second)
				case EB_Moving:
					elevio.SetMotorDirection(elevator.Direction)
				case EB_Idle:
				}
			}
			elevator.SetAllLights()
		case <-doorClose:
			fmt.Println("door close timer timed out")
			if obstructed {
				doorClose = time.After(3 * time.Second)
			} else {
				switch elevator.Behavior {
				case EB_DoorOpen:
					elevator.Direction, elevator.Behavior = NextAction(elevator)
					fmt.Printf("DoorClosed action; \n")
					fmt.Printf("EB: %+v\n", elevator.Behavior)
					fmt.Printf("DIR: %+v\n", elevator.Direction)
					switch elevator.Behavior {
					case EB_DoorOpen:
						doorClose = time.After(3 * time.Second)
						clearAtCurrentFloor(&elevator)
					case EB_Moving:
						elevio.SetMotorDirection(elevator.Direction)
						elevio.SetDoorOpenLamp(false)
					case EB_Idle:
						elevio.SetDoorOpenLamp(false)
						elevio.SetMotorDirection(elevator.Direction)
					}
				}
			}
		case newFloor := <-drv_floors:

			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)
			switch elevator.Behavior {
			case EB_Moving:
				if ShouldStop(elevator) {
					fmt.Printf("STOP\n")
					elevio.SetMotorDirection(elevio.MD_Stop)
					//Opendoor and Start timer
					clearAtCurrentFloor(&elevator)
					elevator.SetAllLights()
					elevio.SetDoorOpenLamp(true)
					doorClose = time.After(3 * time.Second)
					elevator.Behavior = EB_DoorOpen
				}
			case EB_Idle:
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			elevatorChan <- elevator
		case obstructed = <-drv_obstr:
			/* case elevatorChan <- elevator:*/
			/* case elev := <-updateElevatorChan:
			elevator = elev*/
		}
	}
}

func InitializeElevator(elevator *Elevator) {
	elevator.Floor = -1 //Where we are
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Idle
	elevio.SetMotorDirection(elevator.Direction)
}
