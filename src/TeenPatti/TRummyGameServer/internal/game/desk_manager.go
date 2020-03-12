package game

import (
	"fmt"
	// "fmt"
	"time"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"

	"TeenPatti/TRummyGameServer/pkg/errutil"
	"TeenPatti/TRummyGameServer/pkg/room"
	"TeenPatti/TRummyGameServer/protocol"
)

const (
	Offline = "离线"
	Waiting = "等待中"

	fieldDesk   = "desk"
	fieldPlayer = "player"
)

const deskOpBacklog = 64

const (
	errorCode = -1 //错误码
)

const (
	deskNotFoundMessage        = "您输入的房间号不存在, 请确认后再次输入"
	deskPlayerNumEnoughMessage = "您加入的房间已经满人, 请确认房间号后再次确认"
	autherNotAvailMessage      = "无效的帐号认帐，请重新认证"
	versionExpireMessage       = "你当前的游戏版本过老，请更新客户端，地址: http://fir.im/tand"
)

var (
	deskNotFoundResponse = &protocol.JoinDeskResponse{Code: errutil.YXDeskNotFound, Error: deskNotFoundMessage}
	deskPlayerNumEnough  = &protocol.JoinDeskResponse{Code: errorCode, Error: deskPlayerNumEnoughMessage}
	joinVersionExpire    = &protocol.JoinDeskResponse{Code: errorCode, Error: versionExpireMessage}
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
		p.disconnect = true
		// Fixed: 玩家WIFI切换到4G网络不断开, 重连时，将UID设置为illegalSessionUid
		if s.UID() > 0 {

			if err := this.onPlayerDisconnect(s); err != nil {
				logger.Errorf("玩家退出: UID=%d, Error=%s", s.UID, err.Error())
			}
		}
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
func (this *TRDeskManager) getJoinDesk() *Desk {

	for _, desk := range this.desks {

		if desk.IsFullPlayer() == false {
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
	fmt.Println("玩家", data.NickName, "请求加入桌子,Uid:", data.Uid)
	var opts DeskOpts
	//检测消息
	p, err := this.checkSessionAuther(s)

	if err != nil {

		logger.Warnf("JoinDesk Error:", s.UID())

		return s.Response(deskNotAutherSession)
	}

	//得到可加入的桌子
	desk := this.getJoinDesk()

	//若全部桌子都满人了，则建立桌子
	if desk == nil {

		roomNo := room.Next()

		//测试用
		opts.bootAmout = 200
		opts.chaalLimit = 200 * 128
		opts.maxBlinds = 4
		opts.potLimit = 200 * 1024
		opts.betKeepTime = 15

		desk = NewDesk(roomNo, opts)

		//设定桌子
		this.setDesk(roomNo, desk)

	} else {

	}

	//玩家加入桌子
	if err := desk.playerJoin(s, false); err != nil {

		desk.logger.Errorf("玩家加入房间失败，UID=%d, Error=%s", s.UID(), err.Error())
	} else {
		this.playersDesk[s] = desk
	}

	desk.logger.Println("ResponeJoinDesk.........................")
	fmt.Println("session", s, p.session)
	return desk.PlayerJoinAfterInfo(p)

}

//退出桌子
func (this *TRDeskManager) ExitDesk(s *session.Session, msg *protocol.ExitRequest) error {

	return nil
}

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
	return p.desk.OperCard(p, msg)
}

//show请求
func (this *TRDeskManager) ShowCard(s *session.Session, msg *protocol.GShowCardsRequest) error {
	//检测消息
	p, err := this.checkSessionAuther(s)
	if err != nil {
		logger.Warnf("ShowCard Error:", s.UID())
		return s.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: autherNotAvailMessage,
		})
	}
	//检测玩家是否在桌子中
	if p.desk == nil {
		logger.Debug("ShowCard Error:", s.UID())
		return s.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: "玩家没有加入桌子！",
		})
	}
	return p.desk.ShowCards(p, msg)
}
