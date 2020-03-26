package main

import (
	"Robot/TRummyGameRobot/robot"
	"time"
)

func main() {
	RobotStart()
}

func RobotStart() {
	m := robot.NewManager()
	m.Run()
	for {
		time.Sleep(10 * time.Second)
	}
}
