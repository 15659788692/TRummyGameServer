package robot

import (
	"Robot/TRummyGameRobot/io"
	"Robot/TRummyGameRobot/protocol"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const (
	StateFree = iota
	StateLogin
	StateGame
	StateExit
)

type Client struct {
	Connector *io.Connector
	chReady   chan struct{}
	State     int
	//机器人部分

	MyPlayer *Player
}

func NewUser() *Client {
	return &Client{
		Connector: io.NewConnector(),
		chReady:   make(chan struct{}),
		MyPlayer:  &Player{},
		State:     StateFree,
	}

}

func (this *Client) Start(addr string) error {
	this.Connector.OnConnected(func() {
		this.chReady <- struct{}{}
	})
	if err := this.Connector.Start(addr); err != nil {
		fmt.Println("连接错误：", err)
		return err
	}
	<-this.chReady
	time.Sleep(time.Second / 2)
	//游戏逻辑
	go this.Run()
	return nil
}

//游戏流程
func (this *Client) Run() {
	switch this.State {
	case StateFree:
		this.LoginRequest()
	case StateLogin:
		this.JionDeskRequest()
	case StateGame: //游戏逻辑
		this.StartGame()
	}

}

//登录
func (this *Client) LoginRequest() {
	var redis = redis.NewClient(&redis.Options{
		Addr:     "192.168.0.104:6379",
		Password: "",
		DB:       1,
	})
	_, err := redis.Ping().Result()
	if err != nil {
		panic("redis连接失败")
	}
	redis.Get(redis)
	loginServer := &protocol.LoginToGameServerRequest{}
	loginServer.Version = "Version1.0"
	loginServer.Token = "1234"
	loginServer.Uid = 12234344 + int64(time.Now().Second()%100)
	this.Connector.Request(ReqLogin, loginServer, this.LoginResponse)
}

//收到登录回复
func (this *Client) LoginResponse(v interface{}) {
	data := v.([]byte)
	resp := &protocol.LoginToGameServerResponse{}
	json.Unmarshal(data, resp)
	if len(resp.Error) != 0 || !resp.Success {
		fmt.Printf("登录失败！%v\n", resp)
		return
	}

	this.State = StateLogin
	this.MyPlayer.Nickname = resp.Nickname
	this.MyPlayer.HeadUrl = resp.HeadUrl
	this.MyPlayer.Uid = resp.Uid
	this.MyPlayer.Sex = resp.Sex
	fmt.Printf("接收到登陆回复消息：%v\n", resp)
	go this.Run()
}

//加入桌子
func (this *Client) JionDeskRequest() {
	Req := protocol.JoinDeskRequest{}
	Req.NickName = this.MyPlayer.Nickname
	Req.Uid = this.MyPlayer.Uid
	this.Connector.Request(ReqDeskJoinDesk, Req, this.JionDeskResponse)
}

//加入桌子回复
func (this *Client) JionDeskResponse(v interface{}) {
	data := v.([]byte)
	resp := &protocol.JoinDeskResponse{}
	json.Unmarshal(data, resp)
	fmt.Printf("接收到加入桌子回复消息：%v\n", resp)
	if !resp.Success {
		fmt.Printf("加入桌子失败！%v\n", resp)
		return
	}
	this.State = StateGame
	this.MyPlayer.MyDesk = &Desk{
		DeskInfo:  resp.DeskInfo,
		Connector: this.Connector,
	}
	go this.Run()
}

//游戏逻辑
func (this *Client) StartGame() {
	//初始化监听广播和推送
	this.MyPlayer.InitPlayer()
}
