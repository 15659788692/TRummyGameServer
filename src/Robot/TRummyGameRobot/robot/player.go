package robot

import (
	"Robot/TRummyGameRobot/io"
	"Robot/TRummyGameRobot/protocol"
	"encoding/json"
	"time"
)

type Player struct {
	Connector *io.Connector
	Nickname  string
	Uid       int64
	HeadUrl   string
	Sex       int

	MyDesk   *Desk
	HandCard []int32
}

func (this *Player) InitData() {

}

func (this *Player) InitPlayer() {
	this.Connector.On(PushSendCard, this.SendCard)
	this.Connector.On(NoticeGameState, this.GameState)
	this.MyDesk.InitDesk()
}

//发牌
func (this *Player) SendCard(v interface{}) {

	data := v.([]byte)
	resp := &protocol.GSendCardNotice{}
	json.Unmarshal(data, resp)
	this.HandCard = append([]int32{}, resp.HandCards...)
	this.MyDesk.WildCard = resp.WildCard
	this.MyDesk.FristCard = resp.FristCard

}

//删除手中的牌
func (this *Player) DelHandCard(card int32) bool {
	//检测玩家手中是否有手牌
	if len(this.HandCard) <= 1 {
		return false
	}
	//检测手牌中是否有这张牌
	for k, v := range this.HandCard {
		if v == card {
			this.HandCard = append(this.HandCard[:k], this.HandCard[k+1:]...)
			return true
		}
	}
	return false
}

//操作请求
func (this *Player) PlayRequest() {
	time.Sleep(time.Second)
	if len(this.HandCard) == 13 {
		this.Connector.Request("TRDeskManager.OperCard", &protocol.GOperCardRequest{
			Opertion: 2,
		}, this.PlayResponse)
		return
	}
	if len(this.HandCard) == 14 {
		this.Connector.Request("TRDeskManager.OperCard", &protocol.GOperCardRequest{
			Opertion: 3,
			OperCard: this.HandCard[0],
		}, this.PlayResponse)
	}
}

//操作应答
func (this *Player) PlayResponse(v interface{}) {
	data := v.([]byte)
	resp := &protocol.GOperCardResponse{}
	json.Unmarshal(data, resp)

	if len(this.HandCard) == 13 {
		this.HandCard = append(this.HandCard, resp.OperCard)
		time.Sleep(time.Second)
		this.PlayRequest()
		return
	}
	if len(this.HandCard) == 14 {
		this.DelHandCard(resp.OperCard)
		return
	}
}

//玩家状态通知
func (this *Player) GameState(v interface{}) {
	//验证
	//解析
	data := v.([]byte)
	resp := &protocol.GGameStateNotice{}
	json.Unmarshal(data, resp)
	this.MyDesk.DeskInfo.GameState = resp.GameState
	//分状态处理
	if this.MyDesk.DeskInfo.GameState == GameStatePlay {
		this.MyDesk.DeskInfo.OperSeatId = resp.OperateId
		if resp.OperateId == this.MyDesk.DeskInfo.UserSeatId {
			this.PlayRequest()
		}
	}
}
