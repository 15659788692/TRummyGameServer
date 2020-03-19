package main

import (
	"TeenPatti/TPClientTest/io"
	"TeenPatti/TPClientTest/protocol"
	"time"

	"encoding/json"
	"fmt"
)

/*
type syser struct {

	heartbeat float64    `json:"heartbeat"`
}

type header struct {

	code    int   `json:"code"`
	sys    []syser  `json:"sys"`
}
*/

func main() {

	TestClient()

}

//--------------------------------------------------------------------
const (
	addr = "192.168.0.141:3000" // local address

	// addr = "127.0.0.1:3000" // local address
	conc = 1000 // concurrent client count
)

var clientHander *io.Connector

func TestClient() {

	c := io.NewConnector()

	clientHander = c

	chReady := make(chan struct{})

	c.OnConnected(func() {
		chReady <- struct{}{}
	})

	if err := c.Start(addr); err != nil {
		panic(err)
	}

	time.Sleep(time.Second / 2)

	c.On("TRGame.BroadDeskPlayersInfo", OnBroadMessage)

	//发送登陆消息
	loginServer := &protocol.LoginToGameServerRequest{}
	loginServer.Version = "Version1.0"
	loginServer.Token = 123456
	loginServer.Uid = 12234344 + int64(time.Now().Second()%100)

	// time.Sleep(time.Second)
	fmt.Println("1")
	//发送登陆请求
	c.Request("TRManager.Login", loginServer, testLoginResponse)

	fmt.Println("连接：", <-chReady)

	for {
		time.Sleep(10 * time.Millisecond)
	}

}

//----------------------------------------------------------------------------------------------

//处理消息回复1
func testLoginResponse(v interface{}) {

	data := v.([]byte)

	resp := &protocol.LoginToGameServerResponse{}

	fmt.Println("接收到登陆回复消息：")
	json.Unmarshal(data, resp)

	fmt.Println(resp)

	//发送加入桌子请求
	JoinDeskMessage(clientHander, resp.Uid)

}

//----------------------------------------------------------------------------
func JoinDeskMessage(cc *io.Connector, uid int64) {

	req := &protocol.JoinDeskRequest{}
	req.Uid = uid

	cc.Request("TPDeskManager.JoinDesk", req, testJoinDeskResponse)

	fmt.Println("发送加入桌子的消息：")

}

//处理消息回复2
func testJoinDeskResponse(v interface{}) {

	data := v.([]byte)

	resp := &protocol.JoinDeskResponse{}

	json.Unmarshal(data, resp)

	fmt.Println("接收到加入桌子回复消息：")

	fmt.Println(resp)

	//发送客户端资源已准备好的消息
	req := &protocol.ClientInitCompletedNotify{}
	req.IsReEnter = false

	clientHander.Notify("TPDeskManager.ClientInitCompleted", req)

	fmt.Println("发送客户端资源已准备好的消息：")
}

//--------------------------------------------------------------------------------

func OnBroadMessage(v interface{}) {

	fmt.Println("收到广播消息了")

	data := v.([]byte)

	resp := &protocol.PlayerEnterDesk{}

	json.Unmarshal(data, resp)

	fmt.Println("desk playernum:", len(resp.Data))

	for k, v := range resp.Data {

		fmt.Println("player No:", k, v.Nickname, v.IsBanker, v.Sitdown)

	}

}
