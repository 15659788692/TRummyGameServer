package game

import (
	"fmt"
	"github.com/lonng/nano"
	"github.com/lonng/nano/session"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"TeenPatti/TPattiGameServer/pkg/constant"
	"TeenPatti/TPattiGameServer/pkg/errutil"
	"TeenPatti/TPattiGameServer/pkg/room"
	"TeenPatti/TPattiGameServer/protocol"
)

const (
	DeskFullplayerNum = 5 //总的玩家数量
)

type DeskOpts struct {
	bootAmout int //低注,进入此桌最少的投注额
	maxBlinds int //最大盲注,最多可盖牌的圈数

	chaalLimit int //单注限额
	potLimit   int //单局总投注额度

	betKeepTime int //总的投注时间,秒数
}

type Desk struct {
	roomNo room.Number //房间号

	deskID int64 //desk表的pk

	deskOpt DeskOpts //此桌的参数

	//	opts      *protocol.DeskOptions // 房间选项

	state constant.DeskStatus // 状态

	round uint32 // 第n局

	creator   int64 // 创建玩家UID
	createdAt int64 // 创建时间

	playMutex sync.RWMutex

	players []*Player //桌上玩家

	group *nano.Group // 组播通道

	die chan struct{}

	isNewRound   bool //是否是每局的第一次出牌
	isFirstRound bool //是否是本桌的第一局牌

	betTime int64 //每个玩家可以投注的时间

	totalBet int64 //总投注

	latestEnter *protocol.PlayerEnterDesk //最新的进入状态

	logger *log.Entry
}

func NewDesk(roomNo room.Number, opts DeskOpts) *Desk {

	d := &Desk{

		state: constant.DeskStatusCreate,

		roomNo:  roomNo,
		deskOpt: opts,

		players: []*Player{},
		group:   nano.NewGroup(uuid.New()),

		die: make(chan struct{}),

		isNewRound:   true,
		isFirstRound: true,

		logger: log.WithField("deskno", roomNo),
	}

	logger.Println("new desk:", roomNo.String())

	return d
}

// 玩家数量
func (d *Desk) totalPlayerCount() int {

	d.playMutex.Lock()
	defer d.playMutex.Unlock()

	return len(d.players)

}

//检测桌子人数是否满了
func (d *Desk) IsFullPlayer() bool {

	d.playMutex.Lock()
	defer d.playMutex.Unlock()

	if len(d.players) == DeskFullplayerNum {
		return true
	}

	return false
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

//发送桌上玩家的状态信息给 每个人
func (d *Desk) syncDeskStatus() {

	d.latestEnter = &protocol.PlayerEnterDesk{Data: []protocol.EnterDeskInfo{}}

	d.playMutex.Lock()

	for _, p := range d.players {

		//若玩家站起来了，则不发送
		//if p.GetSitdown() == false {
		//	continue
		//}

		d.latestEnter.Data = append(d.latestEnter.Data, protocol.EnterDeskInfo{
			SeatPos:  p.seatPos,
			Nickname: p.name,
			Sex:      p.sex,
			HeadUrl:  p.head,

			Score:   p.betScore,
			StarNum: p.starNum,

			IsBanker: p.isBanker,
			Sitdown:  p.sitdown,
			Betting:  p.IsBetting(),

			Packed: p.packed,
			Show:   p.showed,
			Blind:  p.blinded,
		})
	}

	d.playMutex.Unlock()

	d.logger.Println("BroadDeskPlayersInfo ")

	//通知大家
	d.group.Broadcast("BroadDeskPlayersInfo", d.latestEnter)

}

func (d *Desk) checkStart() {

	s := d.status()

	if (s != constant.DeskStatusCreate) && (s != constant.DeskStatusCleaned) {

		d.logger.Println("启动游戏不成功，状态不符合:", s.String())
		return
	}

	if d.isFirstRound == false {

		return
	}

	//检测加入桌子的人数，至少要2个人游戏才可以开始
	if len(d.players) < 2 {

		d.logger.Println("启动游戏不成功，人数不够:", len(d.players))

		return
	}

	d.logger.Println("启动游戏", s.String())

	d.start()
}

func (d *Desk) title() string {
	return strings.TrimSpace(fmt.Sprintf("房号: %s 局数: %d/%d", d.roomNo, d.round))
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

	d.logger.Debug("开始游戏")

	//发送每个玩家数据给全桌的人
	d.syncDeskStatus()

	//	time.AfterFunc()

	//curPlayer := d.players[d.curTurn] //当前出牌玩家,初始为庄家

	if d.status() != constant.DeskStatusInterruption {
		d.setStatus(constant.DeskStatusRoundOver)
	}

	d.roundOver()
}

// 牌桌准备开始，倒计时5秒动画
func (d *Desk) statusReadyA() {

	d.round++

	d.setStatus(constant.DeskStatusReadyStart)

	tableStatus := &protocol.DeskStatusInfo{

		Status:     (int32)(constant.DeskStatusReadyStart),
		KeepSecond: 5,
	}

	//通知全桌玩家
	d.group.Broadcast(constant.DeskStatusRoute, tableStatus)

}

//进入吸收筹码的状态，吸筹码动画
func (d *Desk) statusReadyB() {

	d.setStatus(constant.DeskStatusReadyLowBet)

	tableStatus := &protocol.DeskStatusInfo{
		Status:     (int32)(constant.DeskStatusReadyStart),
		KeepSecond: 3,
	}

	//通知全桌玩家
	d.group.Broadcast(constant.DeskStatusRoute, tableStatus)

	//每个玩家扣除点数，然后通知玩家

}

//进入发牌动画
func (d *Desk) statusReadyC() {

	d.setStatus(constant.DeskStatusReadyFaiPai)

	tableStatus := &protocol.DeskStatusInfo{
		Status:     (int32)(constant.DeskStatusReadyFaiPai),
		KeepSecond: 5,
	}

	//通知全桌玩家
	d.group.Broadcast(constant.DeskStatusRoute, tableStatus)

}

//得到当前作庄的玩家,相当于第一个出牌的
func (d *Desk) GetBankerPlayer() *Player {

	d.playMutex.Lock()
	defer d.playMutex.Unlock()

	if len(d.players) == 0 {
		return nil
	}

	for _, player := range d.players {

		if player == nil {

			d.logger.Println("GetBankerPlayer errror ")
			continue
		}

		if player.isBanker == true {

			return player
		}
	}

	return nil

}

//读取下一个需要投注的玩家，按顺时间方向
func (d *Desk) GetNextPlayer(player *Player) *Player {

	d.playMutex.Lock()
	defer d.playMutex.Unlock()

	if len(d.players) == 0 {
		return nil
	}

	for _, nextplayer := range d.players {

		if nextplayer.seatPos == player.seatPos {
			return player
		}
	}

	return nil

}

//通知大家，谁当前可投注了
func (d *Desk) BroadPlayerBetActive(nickName string, seatPos int, keepTime int, lostTime int) {

	bmsg := &protocol.BroadPlayerBetAcitve{
		NickName: nickName,
		SeatPos:  seatPos,
		KeepTime: keepTime,
		LostTime: lostTime,
	}

	d.group.Broadcast("BroadPlayersBetting", bmsg)

}

//等待当前玩家的投注
func (d *Desk) waitPlayerBetting(player *Player) error {

	//每个人有15秒的投注时间
	playTick := time.After(time.Duration(d.betTime))

	betSignal := player.GetBetSignal()

	select {

	case <-playTick: //投注时间到
		//需要扣除玩家的点数,同时起立

	case <-betSignal: //检测到有投注了

	case <-d.die: //退出桌子
	}

	return nil

}

//检测桌子的玩家是否可以结算了
func (d *Desk) deskIsAccount() bool {

	return false

}

//进入投注环节
func (d *Desk) statusPlaying() {

	d.setStatus(constant.DeskStatusPlaying)

	tableStatus := &protocol.DeskStatusInfo{
		Status:     (int32)(constant.DeskStatusPlaying),
		KeepSecond: 0,
	}

	//通知全桌玩家
	d.group.Broadcast(constant.DeskStatusRoute, tableStatus)

	for {

		//找出当前当庄的玩家
		player := d.GetBankerPlayer()

		if player == nil {

			logger.Panic("GetBankerPlayer Error ,nil")

			return
		}

		bmsg := &protocol.BroadBankerInfo{

			NickName: player.name,
			SeatPos:  player.seatPos,
		}

		//通知谁当庄家，要第一个出牌
		d.group.Broadcast("BroadCurrentBanker", bmsg)

		//开始投注环节
		for player != nil {

			//通知谁当前可以投注
			d.BroadPlayerBetActive(player.name, player.seatPos, 15, 0)

			//等待此玩家投注完
			err := d.waitPlayerBetting(player)

			if err != nil {

				//玩家起立
				player.SetSitdown(false)

				//通知当前玩家起立

				//通知桌面上其它玩家

			} else {

				//投注的额度加入数据库
			}

			//检测是否可以结算
			if d.deskIsAccount() == true {

				//进入结算环节
				break

			} else {

				//读取下一个玩家
			}

		} //end betting
	}

}

//进入结算环节
func (d *Desk) statusPlayAccount() {

}

func (d *Desk) start() {

	go d.RunDesk()

}

func (d *Desk) RunDesk() {

	var bstop bool

	bstop = false

	d.isFirstRound = false

	d.logger.Println("RunDesk.....")

	for {

		//玩家是否只有1人了
		if d.totalPlayerCount() == 1 {

			d.logger.Println("RunDesk.....只有一个玩家了，退出桌子 ")
			break
		}

		//把桌上的玩家情况，数据同步一下
		d.syncDeskStatus()

		//游戏开始
		d.statusReadyA()

		d.logger.Println("启动动画显示等待...")

		//倒计时5秒，让客户端有时间显示动画
		timeTickA := time.After(5 * time.Second)

		select {
		case <-timeTickA: //倒计时间到
			break
		case <-d.die: //停止游戏
			bstop = true
			break
		}

		if bstop {
			break
		}

		d.logger.Println("启动吸收筹码动画****")

		//进入收筹码低注的动画
		d.statusReadyB()
		timeTickB := time.After(5 * time.Second)

		select {

		case <-timeTickB: //倒计时间到
			break
		case <-d.die: //停止游戏
			bstop = true
			break

		}

		if bstop {
			break
		}

		d.logger.Println("进入发牌动画#####")

		//进入发牌动画的阶段
		d.statusReadyC()

		timeTickC := time.After(5 * time.Second)

		select {

		case <-timeTickC: //倒计时间到
			break
		case <-d.die: //停止游戏
			bstop = true
			break
		}

		if bstop {
			break
		}

		d.logger.Println("进入游戏环节****")

		//进入游戏投注
		//	d.statusPlaying()

		//进入结算环节
		//    d.statusPlayAccount()
	}

	d.isFirstRound = true
	d.state = constant.DeskStatusCleaned

}

//玩家是否为空了
func (d *Desk) PlayersIsEmpty() bool {

	bempty := false

	if d.totalPlayerCount() == 0 {

		bempty = true
	}

	return bempty
}

//桌子是否处于可投注状态
func (d *Desk) DeskIsCanBet() bool {

	return d.state == constant.DeskStatusPlaying
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

//设置桌子的状态
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
	//return 1 << uint(d.opts.MaxFan)

	return 0
}

func (d *Desk) onPlayerExit(s *session.Session, isDisconnect bool) {

	d.logger.Println("玩家下线了：uid", s.UID())

	uid := s.UID()
	d.group.Leave(s)

	if isDisconnect {
		//	d.dissolve.updateOnlineStatus(uid, false)
	} else {

		restPlayers := []*Player{}

		for _, p := range d.players {

			if p.Uid() != uid {
				restPlayers = append(restPlayers, p)
			} else {
				//p.reset()
				p.desk = nil
				//p.score = 1000
				//p.turn = 0
			}
		}
		d.players = restPlayers
	}

	/*//如果桌上已无玩家, destroy it
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

//检测是否超出限额
func (d *Desk) BetLimitCheck(bet int) (bool, error) {

	//检测低注
	if bet < d.deskOpt.bootAmout {
		return false, nil
	}

	//检测高注
	if bet > d.deskOpt.chaalLimit {

		return false, nil
	}

	return true, nil
}

//投注额度
func (d *Desk) PlayerBet(player *Player, bet int) (int, error) {

	betscore, err := player.PlayBet(bet)

	d.totalBet += int64(bet)

	return betscore, err
}

//通知其它玩家投注
func (d *Desk) BroadPlayerBetting(s *session.Session, nickname string, seatpos int, bet int) {

	resp := &protocol.BroadPlayerBetting{}

	resp.BetCount = int32(bet)
	resp.NickName = nickname
	resp.SeatPos = seatpos

	//通知桌子上其它玩家
	d.group.Multicast("TP.BroadPlayerBet", resp, func(sss *session.Session) bool {

		if sss == s {
			return false
		}
		return true
	})

}

//通知指定玩家起立状态
func (d *Desk) NotifyPlayerSitdown(player *Player, sitdown bool) {

	resp := &protocol.NotifyPlayerSitdown{
		NickName: player.name,
		Uid:      player.uid,
		SeatPos:  player.seatPos,
		Sitdown:  sitdown,
	}

	player.session.Push("TP.NotifyPlayerSitdown", resp)

}

//通知大家，指定玩家的起立状态
func (d *Desk) BroadPlayersSitdown(player *Player, sitdown bool) {

	resp := &protocol.BroadsPlayerSitdown{}

	resp.NickName = player.name
	resp.SeatPos = player.seatPos
	resp.Sitdown = sitdown

	//通知桌子上其它玩家
	d.group.Broadcast("TP.BroadPlayerSitdown", resp)
}
