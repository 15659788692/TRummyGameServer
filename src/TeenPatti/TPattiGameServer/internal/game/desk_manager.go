package game

import (
	"TeenPatti/TPattiGameServer/pkg/errutil"
	"TeenPatti/TPattiGameServer/pkg/room"
	"TeenPatti/TPattiGameServer/protocol"
	"fmt"
	"time"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"
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

type (
	TPDeskManager struct {
		component.Base
		//桌子数据
		desks map[room.Number]*Desk // 所有桌子
	}
)

var defaultDeskManager = NewDeskManager()

func NewDeskManager() *TPDeskManager {

	return &TPDeskManager{
		desks: map[room.Number]*Desk{},
	}

}

func (manager *TPDeskManager) AfterInit() {

	session.Lifetime.OnClosed(func(s *session.Session) {

		logger.Println("**************session OnClose :uid ", s.UID())

		// Fixed: 玩家WIFI切换到4G网络不断开, 重连时，将UID设置为illegalSessionUid
		if s.UID() > 0 {

			if err := manager.onPlayerDisconnect(s); err != nil {
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

		manager.dumpDeskInfo()

	})
}

func (manager *TPDeskManager) dumpDeskInfo() {

	c := len(manager.desks)
	if c < 1 {
		return
	}

	logger.Infof("剩余房间数量: %d 在线人数: %d  当前时间: %s", c, defaultManager.sessionCount(), time.Now().Format("2006-01-02 15:04:05"))
	for no, d := range manager.desks {
		logger.Debugf("房号: %s, 创建时间: %s, 创建玩家: %d, 状态: %s, 总局数: %d, 当前局数: %d",
			no, time.Unix(d.createdAt, 0).String(), d.creator, d.status().String(), 1000, d.round)
	}
}

func (manager *TPDeskManager) onPlayerDisconnect(s *session.Session) error {

	fmt.Println("player DisConnect :uid", s.UID())

	defaultManager.offline(s.UID())

	p, err := playerWithSession(s)

	if err == nil {

		p.desk.onPlayerExit(s, false)
	}

	return nil
}

//根据桌号返回牌桌数据
func (manager *TPDeskManager) desk(number room.Number) (*Desk, bool) {

	d, ok := manager.desks[number]

	return d, ok
}

//设置桌号对应的牌桌数据
func (manager *TPDeskManager) setDesk(number room.Number, desk *Desk) {

	if desk == nil {
		delete(manager.desks, number)
		logger.WithField(fieldDesk, number).Debugf("清除房间: 剩余: %d", len(manager.desks))
	} else {
		manager.desks[number] = desk
	}
}

//读取没有满人的桌子
func (manager *TPDeskManager) getJoinDesk() *Desk {

	for _, desk := range manager.desks {

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

func (manager *TPDeskManager) checkSessionAuther(s *session.Session) (*Player, error) {

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
func (manager *TPDeskManager) JoinDesk(s *session.Session, data *protocol.JoinDeskRequest) error {

	var opts DeskOpts

	//检测消息
	_, err := manager.checkSessionAuther(s)

	if err != nil {

		logger.Warnf("JoinDesk Error:", s.UID())

		return s.Response(deskNotAutherSession)
	}

	//得到可加入的桌子
	desk := manager.getJoinDesk()

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
		manager.setDesk(roomNo, desk)

	} else {

	}

	//玩家加入桌子
	if err := desk.playerJoin(s, false); err != nil {

		desk.logger.Errorf("玩家加入房间失败，UID=%d, Error=%s", s.UID(), err.Error())
	}

	desk.logger.Println("ResponeJoinDesk.........................")

	err = s.Response(&protocol.JoinDeskResponse{
		Success: true,
		TableInfo: protocol.TableInfo{
			DeskNo:     desk.roomNo.String(),
			Status:     int32(desk.status()),
			Round:      desk.round,
			MaxBlinds:  opts.maxBlinds,
			ChaalLimit: opts.chaalLimit,
		},
	})

	return err

}

//退出桌子
func (manager *TPDeskManager) ExitDesk(s *session.Session, msg *protocol.ExitRequest) error {

	return nil
}

//玩家进入桌子成功后，需发送此条消息,用于通知其它人加入成功
func (manager *TPDeskManager) ClientInitCompleted(s *session.Session, msg *protocol.ClientInitCompletedNotify) error {

	uid := s.UID()

	p, err := playerWithSession(s)

	if err != nil {
		return err
	}

	logger.Println(" ClientInitCompletedNotify")

	d := p.desk

	// 客户端准备完成后加入消息广播队列
	for _, p := range d.players {

		if p.Uid() == uid {

			if p.session != s {
				p.logger.Error("DeskManager.ClientInitCompleted: Session不一致")
			}

			p.logger.Info("DeskManager.ClientInitCompleted: 玩家加入房间广播列表")

			d.group.Add(p.session)
			break
		}
	}

	//设定玩家坐下
	p.SetSitdown(true)

	// 如果不是重新进入游戏, 则同步状态到房间所有玩家
	if !msg.IsReEnter {
		d.syncDeskStatus()
	}

	//检测人数情况，符合人数就开启开始游戏
	d.checkStart()

	return err
}

//玩家压注
func (manager *TPDeskManager) BlinkPush(s *session.Session, msg *protocol.PlayerBetRequest) error {

	logger.Debug(msg)

	p, err := playerWithSession(s)

	if err != nil {
		return err
	}

	resp := &protocol.PlayerBetResponse{
		Uid:      msg.Uid,
		BetCount: msg.BetCount,
		Success:  false,
	}

	//检测是否属于玩家下注状态
	if p.IsBetting() == false {

		resp.Error = "not bettin status "

		s.Response(resp)

		return nil
	}

	//桌子为空
	if p.desk == nil {

		return s.Response(resp)
	}

	//检测是否符合高低注条件
	ischeck, _ := p.desk.BetLimitCheck((int)(msg.BetCount))

	if ischeck == false {

		return s.Response(resp)
	}

	totalbet, err := p.desk.PlayerBet(p, (int)(msg.BetCount))

	resp.BetCount = msg.BetCount
	resp.TotalBet = int64(totalbet)

	resp.Success = true

	//回复给玩家端
	err = s.Response(resp)

	//广播通知其它玩家,此玩家的下注
	p.desk.BroadPlayerBetting(s, p.name, p.seatPos, int(msg.BetCount))

	return err
}

//按下了pack按钮,需要回复按下后的情况消息
func (manager *TPDeskManager) PackPush(s *session.Session) error {

	if _, err := manager.checkSessionAuther(s); err != nil {

		return err
	}

	return nil
}

//按下了看牌按钮，需要回复给三张牌的内容
func (manager *TPDeskManager) SeePush(s *session.Session, msg *protocol.PlayerSeeRequest) error {

	if _, err := manager.checkSessionAuther(s); err != nil {

		return err
	}

	return nil
}

//按下了显示按钮
func (manager *TPDeskManager) ShowPush(s *session.Session) error {

	if _, err := manager.checkSessionAuther(s); err != nil {

		return err
	}

	return nil
}

//按下了坐下
func (manager *TPDeskManager) SitdownPush(s *session.Session) error {

	if _, err := manager.checkSessionAuther(s); err != nil {

		return err
	}

	return nil

}
