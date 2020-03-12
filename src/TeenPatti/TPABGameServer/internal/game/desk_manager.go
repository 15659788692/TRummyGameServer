package game

import (
	"TeenPatti/TPABGameServer/pkg/errutil"
	"TeenPatti/TPABGameServer/pkg/room"
	"TeenPatti/TPABGameServer/protocol"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"
	"time"
)

const (
	Offline       = "离线"
	Waiting       = "等待中"

	fieldDesk   = "desk"
	fieldPlayer = "player"
)

const deskOpBacklog = 64

const (
	errorCode             = -1  //错误码
)

const (
	deskNotFoundMessage        = "您输入的房间号不存在, 请确认后再次输入"
	deskPlayerNumEnoughMessage = "您加入的房间已经满人, 请确认房间号后再次确认"
	versionExpireMessage       = "你当前的游戏版本过老，请更新客户端，地址: http://fir.im/tand"
)


var (
	deskNotFoundResponse = &protocol.JoinDeskResponse{Code: errutil.YXDeskNotFound, Error: deskNotFoundMessage}
	deskPlayerNumEnough  = &protocol.JoinDeskResponse{Code: errorCode, Error: deskPlayerNumEnoughMessage}
	joinVersionExpire    = &protocol.JoinDeskResponse{Code: errorCode, Error: versionExpireMessage}

)

type (
	DeskManager struct {
		component.Base
		//桌子数据
		desks map[ room.Number ]*Desk // 所有桌子
	}
)

var defaultDeskManager = NewDeskManager()


func NewDeskManager() *DeskManager {

	return &DeskManager{
		desks: map[room.Number]*Desk{},
	}



}

func (manager *DeskManager) AfterInit() {


	session.Lifetime.OnClosed(  func(s *session.Session) {


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

func (manager *DeskManager) dumpDeskInfo() {
	c := len(manager.desks)
	if c < 1 {
		return
	}

	logger.Infof("剩余房间数量: %d 在线人数: %d  当前时间: %s", c, defaultManager.sessionCount(), time.Now().Format("2006-01-02 15:04:05"))
	for no, d := range manager.desks {
		logger.Debugf("房号: %s, 创建时间: %s, 创建玩家: %d, 状态: %s, 总局数: %d, 当前局数: %d",
			no, time.Unix(d.createdAt, 0).String(), d.creator, d.status().String(), d.opts.MaxRound, d.round)
	}
}

func (manager *DeskManager) onPlayerDisconnect(s *session.Session) error {

	return nil
}


func (Manager *DeskManager ) createDesk(   opts *protocol.DeskOptions ) {

	  roomNo :=room.Next()

	  logger.Info("room No:" , roomNo.String(), opts  )

	  desk :=NewDesk( roomNo, opts )

     //设定桌子
	  Manager.setDesk( roomNo, desk )



}



// 根据桌号返回牌桌数据
func (manager *DeskManager) desk(number room.Number) (*Desk, bool) {


	d, ok := manager.desks[number]

	return d, ok
}

// 设置桌号对应的牌桌数据
func (manager *DeskManager) setDesk(number room.Number, desk *Desk) {

	if desk == nil {
		delete(manager.desks, number)
		logger.WithField(fieldDesk, number).Debugf("清除房间: 剩余: %d", len(manager.desks))
	} else {
		manager.desks[number] = desk
	}

}







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



//进入桌子
func (manager *DeskManager) JoinDesk(s *session.Session, data *protocol.JoinDeskRequest) error {


	if forceUpdate && data.Version != version {
		return s.Response(joinVersionExpire)
	}

    //



	//dn := room.Number(data.DeskNo)

	d, ok := manager.desk(dn)


	if !ok {
		return s.Response(deskNotFoundResponse)
	}

	if len(d.players) >= d.totalPlayerCount() {
		return s.Response(deskPlayerNumEnough)
	}


	if err := d.playerJoin(s, false); err != nil {
		d.logger.Errorf("玩家加入房间失败，UID=%d, Error=%s", s.UID(), err.Error())
	}

	return s.Response(&protocol.JoinDeskResponse{
		TableInfo: protocol.TableInfo{
			DeskNo:    d.roomNo.String(),
			CreatedAt: d.createdAt,
			Creator:   d.creator,
			Title:     d.title(),
			Status:    d.status(),
			Round:     d.round,
			Mode:      d.opts.Mode,
		},
	})

}

//退出桌子
func (manager *DeskManager) ExitDesk(s *session.Session, msg *protocol.ExitRequest) error {


	return nil
}



