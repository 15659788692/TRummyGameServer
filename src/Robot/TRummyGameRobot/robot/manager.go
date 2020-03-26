package robot

import "time"

func NewManager() *manager {
	return &manager{
		Addr: "192.168.0.141:3000",
	}
}

type manager struct {
	Robot []*User
	Addr  string
}

func (this *manager) Run() {
	for i := 0; i < 10; i++ {
		this.AddClient()
		time.Sleep(time.Second)
	}
}

func (this *manager) AddClient() bool {
	user := NewUser()
	user.Start(this.Addr)
	return true
}
