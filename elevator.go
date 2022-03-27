package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

type ElevatorBehavior int

const (
	EB_Idle       ElevatorBehavior = 0
	EB_DoorOpen   ElevatorBehavior = 1
	EB_Moving     ElevatorBehavior = 2
	EB_Initialize ElevatorBehavior = 3
	EB_MotorStop  ElevatorBehavior = 4
	EB_DoorJam	  ElevatorBehavior = 5
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
	elevatorUpdateChan chan<- Elevator,
	getElevChan chan<- Elevator,
	elevatorInitializedChan chan<- bool) {
	var elevator Elevator
	obstructed := false
	//timerFinishedChannel := make(chan int)
	InitializeElevator(&elevator)

	var doorClose <-chan time.Time
	var assumeMotorStop <-chan time.Time
	var assumeDoorJam <-chan time.Time
	for {
		select {

		case btn := <-newOrderChan:
			fmt.Println("NEW ORDER")
			//fmt.Printf("COST: %+v\n", CalculateCost(elevator, btn))
			btnFloor := btn.Floor
			btnType := btn.Button
			 //14 seconds because we assume we won't have any motorstops or door stuck scenarios when we are in only local
			switch elevator.Behavior {
			case EB_DoorOpen,EB_DoorJam:
				if ShouldClearImmediately(elevator, btnFloor, btnType) {
					//START TIMER
					doorClose = time.After(3 * time.Second)
					assumeDoorJam=time.After(10*time.Second)
				} else {
					elevator.AddOrder(btnFloor, btnType)
				}
			case EB_Moving:
				elevator.AddOrder(btnFloor, btnType)
				//assumeMotorStop = time.After(14* time.Second)
			case EB_Idle, EB_MotorStop:
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
					assumeDoorJam=time.After(10*time.Second)
				case EB_Moving:
					elevio.SetMotorDirection(elevator.Direction)
					assumeMotorStop = time.After(14* time.Second)
				}
				
			}
			//elevatorUpdateChan<-elevator WHYYYYY
			elevator.SetAllLights()
		case <-doorClose:
			fmt.Println("door close timer timed out")
			if obstructed {
				doorClose = time.After(3 * time.Second)
			} else {
				switch elevator.Behavior {
				case EB_DoorOpen,EB_DoorJam:
					assumeDoorJam = nil
					elevator.Direction, elevator.Behavior = NextAction(elevator)
					fmt.Printf("DoorClosed action; \n")
					fmt.Printf("EB: %+v\n", elevator.Behavior)
					fmt.Printf("DIR: %+v\n", elevator.Direction)
					switch elevator.Behavior {
					case EB_DoorOpen:
						doorClose = time.After(3 * time.Second)
						clearAtCurrentFloor(&elevator)
						assumeDoorJam=time.After(10*time.Second)
					case EB_Moving:
						elevio.SetMotorDirection(elevator.Direction)
						elevio.SetDoorOpenLamp(false)
						assumeMotorStop = time.After(14* time.Second)
					case EB_Idle:
						elevio.SetDoorOpenLamp(false)
						elevio.SetMotorDirection(elevator.Direction)
					}
				}
				elevatorUpdateChan <- elevator
				
			}

		case newFloor := <-drv_floors:
			fmt.Println("NEW FLOOR")
			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)
			switch elevator.Behavior {
			case EB_Moving,EB_MotorStop:
				if ShouldStop(elevator) {
					fmt.Printf("STOP\n")
					elevio.SetMotorDirection(elevio.MD_Stop)
					//Opendoor and Start timer
					clearAtCurrentFloor(&elevator)
					elevator.SetAllLights()
					elevio.SetDoorOpenLamp(true)
					doorClose = time.After(3 * time.Second)
					assumeDoorJam=time.After(10*time.Second)
					elevator.Behavior = EB_DoorOpen
					elevatorUpdateChan <- elevator
					assumeMotorStop = nil
				}
			case EB_Initialize:
				fmt.Printf("STOP INIT\n")
				elevio.SetMotorDirection(elevio.MD_Stop)
				elevator.Behavior = EB_Idle
				elevatorInitializedChan <- true
			}
			

		case obstructed = <-drv_obstr:
			fmt.Println("Obstructed: ",obstructed)
		case getElevChan <- elevator:
			fmt.Println("sendt elevator")
		case <-assumeMotorStop:
			fmt.Println("I am motorstop plz help")
			elevator.Behavior=EB_MotorStop
			elevatorUpdateChan <- elevator
		case <-assumeDoorJam:
			fmt.Println("My door is stuck plz help")
			elevator.Behavior=EB_DoorJam
			elevatorUpdateChan <- elevator
		}
	}
}

func InitializeElevator(elevator *Elevator) {
	elevator.Floor = 0 //Where we are
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Initialize
	for f := 0; f < NUM_FLOORS; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			elevator.RemoveOrder(f, elevio.ButtonType(btn))
		}
	}
	elevator.SetAllLights()
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)
}
