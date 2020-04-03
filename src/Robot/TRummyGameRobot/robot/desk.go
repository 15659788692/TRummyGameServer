package robot

import (
	"Robot/TRummyGameRobot/io"
	"Robot/TRummyGameRobot/protocol"
	"encoding/json"
)

type Desk struct {
	DeskInfo  protocol.DeskInfo
	Connector *io.Connector
	//
	WildCard  int32
	FristCard int32
	Players   []*Player
}

func NewDesk(p *Player, deskInfo protocol.DeskInfo, connector *io.Connector) (desk *Desk) {
	desk = &Desk{
		DeskInfo:  deskInfo,
		Connector: connector,
	}

	for _, v := range deskInfo.PlayersInfo {
		if v.SeatPos == deskInfo.UserSeatId {
			desk.Players = append(desk.Players, p)
		} else {
			player := &Player{}
			desk.Players = append(desk.Players, player)
		}

	}
	return
}

//初始化监听
func (this *Desk) InitDesk() {
	this.Connector.On(NoticePlayerJoin, this.PlayerJoin)
	this.Connector.On(PushGameStrat, func(interface{}) {})

}

//玩家加入通知
func (this *Desk) PlayerJoin(v interface{}) {
	//验证
	//解析
	data := v.([]byte)
	resp := &protocol.EnterDeskInfo{}
	json.Unmarshal(data, resp)
	if resp.SeatPos == this.DeskInfo.UserSeatId {
		return
	}
	this.DeskInfo.PlayersInfo = append(this.DeskInfo.PlayersInfo, *resp)
}
