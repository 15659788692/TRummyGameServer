package game

import (
	"time"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"

	"TeenPatti/TRummyGameServer/pkg/room"
	"TeenPatti/TRummyGameServer/protocol"
)

const (
	fieldDesk = "desk"
)

const (
	errorCode = -1 //错误码
)

const (
	autherNotAvailMessage = "无效的帐号认帐，请重新认证"
)

var (
	deskNotAutherSession = &protocol.JoinDeskResponse{Code: errorCode, Error: autherNotAvailMessage}
)

type TRDeskManager struct {
	component.Base
	//桌子数据
	desks       map[room.Number]*Desk      // 所有桌子
	playersDesk map[*session.Session]*Desk //玩家所在桌子
}

var defaultDeskManager = NewDeskManager()

func NewDeskManager() *TRDeskManager {

	return &TRDeskManager{
		desks:       map[room.Number]*Desk{},
		playersDesk: map[*session.Session]*Desk{},
	}

}

func (this *TRDeskManager) AfterInit() {

	session.Lifetime.OnClosed(func(s *session.Session) {
		logger.Println("**************session OnClose :uid ", s.UID())
		p, _ := this.checkSessionAuther(s)
		if p.isJoin {
			p.disconnect = true
			p.desk.ExitDesk(p)
		} else { //退出
			desk := p.desk
			desk.Mutex.Lock()
			desk.ExitDesk(p)
			if desk.PlayersIsEmpty() {
				delete(this.desks, desk.roomNo)
			}
			desk.Mutex.Unlock()
			p.ExitDesk()
		}
		log.Println("玩家断线！", p)
	})

	// 每5分钟清空一次已摧毁的房间信息
	scheduler.NewTimer(300*time.Second, func() {

		destroyDesk := map[room.Number]*Desk{}

		//		deadline := time.Now().Add(-24 * time.Hour).Unix()

		//for no, d := range manager.desks {
		// 清除创建超过24小时的房间
		//	if d.status() == constant.DeskStatusDestory || d.createdAt < deadline {
		//		destroyDesk[no] = d
		//	}
		//}

		for _, d := range destroyDesk {
			d.destroy()
		}

		this.dumpDeskInfo()

	})
}

func (this *TRDeskManager) dumpDeskInfo() {

	c := len(this.desks)
	if c < 1 {
		return
	}

	logger.Infof("剩余房间数量: %d 在线人数: %d  当前时间: %s", c, defaultManager.sessionCount(), time.Now().Format("2006-01-02 15:04:05"))
	for no, d := range this.desks {
		logger.Debugf("房号: %s, 创建时间: %s, 创建玩家: %d, 状态: %s, 总局数: %d, 当前局数: %d",
			no, time.Unix(d.createdAt, 0).String(), d.creator, d.gameState, 1000, d.round)
	}
}

func (this *TRDeskManager) onPlayerDisconnect(s *session.Session) error {

	logger.Println("player DisConnect :uid", s.UID())

	defaultManager.offline(s.UID())

	p, err := playerWithSession(s)

	if err == nil {

		p.desk.onPlayerExit(s, false)
	}

	return nil
}

//根据桌号返回牌桌数据
func (this *TRDeskManager) desk(number room.Number) (*Desk, bool) {

	d, ok := this.desks[number]

	return d, ok
}

//设置桌号对应的牌桌数据
func (this *TRDeskManager) setDesk(number room.Number, desk *Desk) {

	if desk == nil {
		delete(this.desks, number)
		logger.WithField(fieldDesk, number).Debugf("清除房间: 剩余: %d", len(this.desks))
	} else {
		this.desks[number] = desk
	}
}

//读取没有满人的桌子
func (this *TRDeskManager) getJoinDesk(p *Player) *Desk {
	if p.desk != nil {
		return p.desk
	}

	for _, desk := range this.desks {

		if desk.IsFullPlayer() == false && desk.gameState <= GameStateWaitStart {
			return desk
		}
	}

	return nil
}

/*


// 网络断开后, 如果ReConnect后发现当前正在房间中, 则重新进入, 桌号是之前的桌号
func (manager *DeskManager) ReJoin(s *session.Session, data *protocol.ReJoinDeskRequest) error {

	d, ok := manager.desk(room.Number(data.DeskNo))

	if !ok || d.isDestroy() {
		return s.Response(&protocol.ReJoinDeskResponse{
			Code:  -1,
			Error: "房间已解散",
		})
	}
	d.logger.Debugf("玩家重新加入房间: UID=%d, Data=%+v", s.UID(), data)

	return d.onPlayerReJoin(s)
}


// 应用退出后重新进入房间
func (manager *DeskManager) ReEnter(s *session.Session, msg *protocol.ReEnterDeskRequest) error {

	return nil
}
*/

func (this *TRDeskManager) checkSessionAuther(s *session.Session) (*Player, error) {

	p, err := playerWithSession(s)

	if err != nil {
		return nil, err
	}

	//检测得到是否空，若空表示之前还没有绑定
	if p == nil {
		return nil, err
	}

	return p, nil

}

//进入桌子
func (this *TRDeskManager) JoinDesk(s *session.Session, data *protocol.JoinDeskRequest) error {
	logger.Println("收到加入桌子请求", data)
	//检测消息
	p, err := this.checkSessionAuther(s)

	if err != nil {

		logger.Warnf("JoinDesk Error:", s.UID())

		return s.Response(deskNotAutherSession)
	}
	isReJoin := false
	if p.desk != nil {
		isReJoin = true
	}
	//得到可加入的桌子
	desk := this.getJoinDesk(p)

	//若全部桌子都满人了，则建立桌子
	if desk == nil {

		roomNo := room.Next()

		desk = NewDesk(roomNo)

		//设定桌子
		this.setDesk(roomNo, desk)

	}

	//玩家加入桌子
	if err := desk.playerJoin(s, isReJoin); err != nil {

		log.Errorf("玩家加入房间失败，UID=%d, Error=%s", s.UID(), err.Error())
	} else {
		this.playersDesk[s] = desk
	}

	//desk.logger.Println("ResponeJoinDesk.........................")
	log.Println(p.name, "加入房间！")
	desk.Mutex.Lock()
	defer desk.Mutex.Unlock()
	return desk.PlayerJoinAfterInfo(p, isReJoin)
}

////退出桌子
//func (this *TRDeskManager) ExitDesk(s *session.Session, msg *protocol.ExitRequest) error {
//
//	return nil
//}

//按下了坐下
func (this *TRDeskManager) SitdownPush(s *session.Session) error {

	if _, err := this.checkSessionAuther(s); err != nil {

		return err
	}

	return nil

}

//游戏部分
//玩家操作牌
func (this *TRDeskManager) OperCard(s *session.Session, msg *protocol.GOperCardRequest) error {
	logger.Println("收到操作请求", msg)
	//检测消息
	p, err := this.checkSessionAuther(s)
	if err != nil {
		logger.Warnf("OperCard Error:", s.UID())
		return s.Response(&protocol.GOperCardResponse{
			Opertion: -1,
			Error:    autherNotAvailMessage,
		})
	}
	//检测玩家是否在桌子中
	if p.desk == nil {
		logger.Debug("OperCard Error:", s.UID())
		return s.Response(&protocol.GOperCardResponse{
			Opertion: 0,
			Error:    "玩家没有加入桌子！",
		})
	}
	p.desk.Mutex.Lock()
	defer p.desk.Mutex.Unlock()
	return p.desk.OperCard(p, msg, false)
}

//show请求
func (this *TRDeskManager) ShowCard(s *session.Session, msg *protocol.GSetHandCardRequest) error {
	logger.Println("收到组牌请求", msg)
	//检测消息
	p, err := this.checkSessionAuther(s)
	if err != nil {
		logger.Warnf("ShowCard Error:", s.UID())
		return s.Response(&protocol.GSetHandCardResponse{
			Success: false,
			Error:   autherNotAvailMessage,
		})
	}
	//检测玩家是否在桌子中
	if p.desk == nil {
		logger.Debug("ShowCard Error:", s.UID())
		return s.Response(&protocol.GSetHandCardResponse{
			Success: false,
			Error:   "玩家没有加入桌子！",
		})
	}
	p.desk.Mutex.Lock()
	defer p.desk.Mutex.Unlock()
	return p.desk.ShowCards(p, msg)
}

func (this *TRDeskManager) Settle(s *session.Session, msg *protocol.GSettleRequect) error {
	logger.Println("收到结算请求", msg)
	//检测消息
	p, err := this.checkSessionAuther(s)
	if err != nil {
		logger.Warnf("ShowCard Error:", s.UID())
		return s.Response(&protocol.GSettleResponse{
			LoseCoins: 0,
			Error:     autherNotAvailMessage,
		})
	}
	//检测玩家是否在桌子中
	if p.desk == nil {
		logger.Debug("ShowCard Error:", s.UID())
		return s.Response(&protocol.GSettleResponse{
			LoseCoins: 0,
			Error:     "玩家没有加入桌子！",
		})
	}
	p.desk.Mutex.Lock()
	defer p.desk.Mutex.Unlock()
	return p.desk.Settle(p, msg, false)
}

//放弃
func (this *TRDeskManager) GiveUp(s *session.Session, msg *protocol.GGiveUpRequect) error {
	logger.Println("收到放弃请求", msg)
	//检测消息
	p, err := this.checkSessionAuther(s)
	if err != nil {
		logger.Warnf("ShowCard Error:", s.UID())
		return s.Response(&protocol.GGiveUpResponse{
			Success: false,
		})
	}
	//检测玩家是否在桌子中
	if p.desk == nil {
		logger.Debug("ShowCard Error:", s.UID())
		return s.Response(&protocol.GGiveUpResponse{
			Success: false,
		})
	}
	log.Println("玩家请求弃牌！")
	p.desk.Mutex.Lock()
	defer p.desk.Mutex.Unlock()
	return p.desk.GiveUp(p, false, msg)
}

//请求出牌记录
func (this *TRDeskManager) OutCardRecord(s *session.Session, msg *protocol.GOutCardRecordRequect) error {
	logger.Println("收到出牌记录请求", msg)
	//检测消息
	p, err := this.checkSessionAuther(s)
	if err != nil {
		logger.Warnf("ShowCard Error:", s.UID())
		return s.Response(&protocol.GOutCardRecordResponse{
			Success: false,
		})
	}
	//检测玩家是否在桌子中
	if p.desk == nil {
		logger.Debug("ShowCard Error:", s.UID())
		return s.Response(&protocol.GOutCardRecordResponse{
			Success: false,
		})
	}
	p.desk.Mutex.Lock()
	defer p.desk.Mutex.Unlock()
	return p.desk.OutCardRecord(p)
}
