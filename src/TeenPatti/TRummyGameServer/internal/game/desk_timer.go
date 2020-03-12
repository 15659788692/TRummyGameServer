package game

import (
	"time"
)

type Timer struct {
	Id int
	H  func(interface{})
	T  int //定时时间
	D  interface{}
}

//定时器
func (this *Desk) DoTimer() {
	for {
		if len(this.TList) == 0 {
			continue
		}
		nlist := []*Timer{}
		olist := []*Timer{}
		for _, v := range this.TList {
			v.T--
			if v.T <= 0 {
				olist = append(olist, v)
			} else {
				nlist = append(nlist, v)
			}
		}
		this.TList = nlist
		for _, v := range olist {
			v.H(v.D)
		}
		time.Sleep(time.Duration(1) * time.Second)
	}

}

func (this *Desk) AddTimer(id int, t int, h func(interface{}), d interface{}) {
	this.TList = append(this.TList, &Timer{
		Id: id,
		H:  h,
		T:  t,
		D:  d,
	})
}

//同一id的定时器只能存在一个
func (this *Desk) AddUniueTimer(id int, t int, h func(interface{}), d interface{}) {
	for i := len(this.TList) - 1; i >= 0; i-- {
		if this.TList[i].Id == id {
			this.TList = append(this.TList[:i], this.TList[i+1:]...)
		}
	}
	this.AddTimer(id, t, h, d)
}

func (this *Desk) DelTimer(id int) {
	for i, v := range this.TList {
		if v.Id == id {
			this.TList = append(this.TList[:i], this.TList[i+1:]...)
			break
		}
	}
}

func (this *Desk) GetTimerNum(id int) int {
	for _, v := range this.TList {
		if v.Id == id {
			return v.T
		}
	}
	return 0
}

//清空定时器
func (this *Desk) ClearTimer() {
	this.TList = []*Timer{}
}
