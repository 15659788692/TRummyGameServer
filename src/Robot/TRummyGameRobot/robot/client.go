package robot

import (
	"Robot/TRummyGameRobot/io"
	"time"
)

type User struct {
	client  *io.Connector
	chReady chan struct{}
}

func NewUser() *User {
	return &User{
		client:  io.NewConnector(),
		chReady: make(chan struct{}),
	}

}

func (this *User) Start(addr string) error {
	this.client.OnConnected(func() {
		this.chReady <- struct{}{}
	})
	if err := this.client.Start(addr); err != nil {
		return err
	}
	<-this.chReady
	time.Sleep(time.Second / 2)
	//游戏逻辑

	return nil
}
