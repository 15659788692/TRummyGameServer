package robot

import "time"

func NewManager() *manager {
	return &manager{
		Addr: "ws://192.168.0.141:3000/",
	}
}

type manager struct {
	Robot []*Client
	Addr  string
}

func (this *manager) Run() {
	for i := 0; i < 2; i++ {
		this.Robot = append(this.Robot, this.AddClient())
		time.Sleep(time.Second)
	}
}

func (this *manager) AddClient() *Client {
	client := NewUser()
	client.Start(this.Addr)
	return client
}
