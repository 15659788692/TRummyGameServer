package game

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/lonng/nano"
	"github.com/lonng/nano/session"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	"TeenPatti/TPABGameServer/pkg/constant"
	"TeenPatti/TPABGameServer/pkg/errutil"
	"TeenPatti/TPABGameServer/pkg/room"
	"TeenPatti/TPABGameServer/protocol"
)

const (
	ResultIllegal = 0

)

const (
	illegalTurn   = -1
	deskDissolved = -1 // 桌子解散标记
	illegalTile   = -1
)

type Desk struct {

	roomNo    room.Number           // 房间号
	deskID    int64                 // desk表的pk

	opts      *protocol.DeskOptions // 房间选项

	state     constant.DeskStatus   // 状态

	round     uint32                // 第n局

	creator   int64                 // 创建玩家UID
	createdAt int64                 // 创建时间


	players   []*Player
	group     *nano.Group 			// 组播通道

	die       chan struct{}


	isNewRound    bool            //是否是每局的第一次出牌
	isFirstRound  bool            //是否是本桌的第一局牌


	logger *log.Entry
}

func NewDesk(roomNo room.Number, opts *protocol.DeskOptions  ) *Desk {

	d := &Desk{

		state:   constant.DeskStatusCreate,
		roomNo:  roomNo,
		players: []*Player{},
		group:   nano.NewGroup(uuid.New()),

		die:     make(chan struct{}),

		isNewRound:   true,
		isFirstRound: true,
			opts:         opts,

		logger: log.WithField("deskno", roomNo),
	}

	return d
}

// 玩家数量
func (d *Desk) totalPlayerCount() int {
	return d.opts.Mode
}


func (d *Desk) save() error {


	return nil
}

// 如果是重新进入 isReJoin: true
func (d *Desk) playerJoin(s *session.Session, isReJoin bool) error {

	uid := s.UID()
	var (
		p   *Player
		err error
	)

	if isReJoin {
		//d.dissolve.updateOnlineStatus(uid, true)

		p, err = d.playerWithId(uid)

		if err != nil {
			d.logger.Errorf("玩家: %d重新加入房间, 但是没有找到玩家在房间中的数据", uid)
			return err
		}

		// 加入分组
		d.group.Add(s)

	} else {
		exists := false
		for _, p := range d.players {
			if p.Uid() == uid {
				exists = true
				p.logger.Warn("玩家已经在房间中")
				break
			}
		}
		if !exists {
			p = s.Value(kCurPlayer).(*Player)
			d.players = append(d.players, p)
			for i, p := range d.players {
				p.setDesk(d, i)
			}
			//d.roundStats[uid] = &history.Record{}
		}
	}

	return nil
}

func (d *Desk) syncDeskStatus() {


}

func (d *Desk) checkStart() {


	s := d.status()

	if (s != constant.DeskStatusCreate) && (s != constant.DeskStatusCleaned) {
		d.logger.Infof("当前房间状态不对，不能开始游戏，当前状态=%s", s.String())
		return
	}

	if count, num := len(d.players), d.totalPlayerCount(); count < num {
		d.logger.Infof("当前房间玩家数量不足，不能开始游戏，当前玩家=%d, 最低数量=%d", count, num)
		return
	}

	/*
	for _, p := range d.players {
		if uid := p.Uid(); !d.prepare.isReady(uid) {
			p.logger.Info("玩家未准备")
			return
		}
	}*/



	d.start()

}

func (d *Desk) title() string {
	return strings.TrimSpace(fmt.Sprintf("房号: %s 局数: %d/%d", d.roomNo, d.round, d.opts.MaxRound))
}


// 牌桌开始, 此方法只在开桌时执行, 非并行
func (d *Desk) start() {


	d.round++
	d.setStatus(constant.DeskStatusDuanPai)

	var (
	//	totalPlayerCount = d.totalPlayerCount() // 玩家数量
	//	totalTileCount   = d.totalTileCount()   // 麻将数量
	)


}


func (d *Desk) isRoundOver() bool {

	//中/终断表示本局结束
	s := d.status()
	if s == constant.DeskStatusInterruption || s == constant.DeskStatusDestory {
		return true
	}

	return true
}

// 循环中的核心逻辑
func (d *Desk) play() {

	defer func() {

		if err := recover(); err != nil {
			d.logger.Errorf("Error=%v", err)

		}
	}()

	d.setStatus(constant.DeskStatusPlaying)
	d.logger.Debug("开始游戏")

	//curPlayer := d.players[d.curTurn] //当前出牌玩家,初始为庄家


	if d.status() == constant.DeskStatusDestory {
		d.logger.Info("已经销毁(三人都离线或解散)")
		return
	}

	if d.status() != constant.DeskStatusInterruption {
		d.setStatus(constant.DeskStatusRoundOver)
	}

	d.roundOver()
}

func (d *Desk) currentPlayer() *Player {
	return nil


}

func (d *Desk) allLosers(win *Player) []int64 {


	loser := []int64{}


	for _, u := range d.players {
		uid := u.Uid()
		//跳过自己与已和玩家
		//if uid == win.Uid() || d.wonPlayers[uid] {
		//	continue
		//}
		loser = append(loser, uid)
	}
	return loser
}


func (d *Desk) setStatus(s constant.DeskStatus) {
	atomic.StoreInt32((*int32)(&d.state), int32(s))
}

func (d *Desk) status() constant.DeskStatus {
	return constant.DeskStatus(atomic.LoadInt32((*int32)(&d.state)))
}

func (d *Desk) roundOver() {

}

func (d *Desk) clean() {

}


func (d *Desk) isDestroy() bool {
	return d.status() == constant.DeskStatusDestory
}

// 摧毁桌子
func (d *Desk) destroy() {

	//删除桌子
	//scheduler.PushTask(func() {
	//	defaultDeskManager.setDesk(d.roomNo, nil)
	//})


}


func (d *Desk) maxScore() int {
	return 1 << uint(d.opts.MaxFan)
}




func (d *Desk) onPlayerExit(s *session.Session, isDisconnect bool) {

	/*
	uid := s.UID()
	d.group.Leave(s)
	if isDisconnect {
		d.dissolve.updateOnlineStatus(uid, false)
	} else {
		restPlayers := []*Player{}
		for _, p := range d.players {
			if p.Uid() != uid {
				restPlayers = append(restPlayers, p)
			} else {
				p.reset()
				p.desk = nil
				p.score = 1000
				p.turn = 0
			}
		}
		d.players = restPlayers
	}

	//如果桌上已无玩家, destroy it
	if d.creator == uid && !isDisconnect {
		//if d.dissolve.offlineCount() == len(d.players) || (d.creator == uid && !isDisconnect) {
		d.logger.Info("所有玩家下线或房主主动解散房间")
		if d.dissolve.isDissolving() {
			d.dissolve.stop()
		}
		d.destroy()

		// 数据库异步更新
		async.Run(func() {
			desk := &model.Desk{
				Id:    d.deskID,
				Round: 0,
			}
			if err := db.UpdateDesk(desk); err != nil {
				log.Error(err)
			}
		})
	}



	 */
}

func (d *Desk) playerWithId(uid int64) (*Player, error) {

	for _, p := range d.players {
		if p.Uid() == uid {
			return p, nil
		}
	}

	return nil, errutil.ErrPlayerNotFound
}

func (d *Desk) setNextRoundBanker(uid int64, override bool) {


}



func (d *Desk) onPlayerReJoin(s *session.Session) error {

/*	// 同步房间基本信息
	basic := &protocol.DeskBasicInfo{
		DeskID: d.roomNo.String(),
		Title:  d.title(),
		Desc:   d.desc(true),
	}

 */

/*
	if err := s.Push("onDeskBasicInfo", basic); err != nil {
		log.Error(err.Error())
		return err
	}

	if err := s.Push("onPlayerEnter", enter); err != nil {
		log.Error(err.Error())
		return err
	}

	p, err := playerWithSession(s)
	if err != nil {
		log.Error(err)
		return err
	}


 */
	if err := d.playerJoin(s, true); err != nil {
		log.Error(err)
	}


	return nil
}


func (d *Desk) loseCoin() {


}
