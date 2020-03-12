package game

import (
	"TeenPatti/TRummyGameServer/Poker"
	"TeenPatti/TRummyGameServer/pkg/errutil"
	"fmt"
	"math/rand"
	"sync"

	// "sync/atomic"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/session"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	// "TeenPatti/TRummyGameServer/pkg/constant"
	"TeenPatti/TRummyGameServer/pkg/room"
	"TeenPatti/TRummyGameServer/protocol"
)

const (
	DeskFullplayerNum   = 5  //总的玩家数量
	DeskMinplayerNum    = 2  //游戏开始的最小人数
	Desk1RoundLosePoint = 20 //第一轮输的点数
	Desk2RoundLosePoint = 40 //第二轮输的点数
	DeskMaxLosePoint    = 80 //最多输80点
)

const ( //游戏状态
	GameStateWaitJoin  = 0
	GameStateWaitStart = 1
	GameStateSendCard  = 2
	GameStatePlay      = 3 //玩家操作阶段
	GameStateStettle   = 4
)
const ( //状态时间
	GameStateWaitStartTime = 7  //游戏开始倒计时
	GameStateSendCardTime  = 9  //发牌
	GameStatePlayTime      = 30 //每个玩家操作时间
	GameStateStettleTime   = 20 //结算时间
)

type DeskOpts struct {
	bootAmout   int //低注,进入此桌最少的投注额
	maxBlinds   int //最大盲注,最多可盖牌的圈数
	chaalLimit  int //单注限额
	potLimit    int //单局总投注额度
	betKeepTime int //总的投注时间,秒数

	//

}

type Desk struct {
	roomNo  room.Number //房间号
	deskID  int64       //desk表的pk
	deskOpt DeskOpts    //此桌的参数
	//	opts      *protocol.DeskOptions // 房间选项
	round     uint32 // 第n轮
	creator   int64  // 创建玩家UID
	createdAt int64  // 创建时间

	playMutex       sync.RWMutex
	players         []*Player         //房间内所有玩家
	seatPlayers     map[int32]*Player //座位上的玩家
	doingPlayers    map[int32]*Player //正在玩的玩家
	group           *nano.Group       //房间内组播通道
	InDeskPlayGroup *nano.Group       //正在玩的玩家广播通道
	isFirstRound    bool              //是否是本桌的第一局牌
	totalBet        int64             //总投注
	logger          *log.Entry

	//游戏部分
	CardMgr       GMgrCard //卡牌管理器
	gameState     int      //游戏状态
	gameStateTime int32    //状态时间
	TList         []*Timer // 定时器列表

	BankerId    int32   //庄家
	FristOutId  int32   //首出玩家
	WildCard    GCard   //万能牌(除大小王外的另一张万能牌)
	ShowCard    GCard   //show的牌
	PublicCard  []GCard //公摊牌堆
	OperateId   int32   //当前正在操作的玩家
	PointValue  int64   //底注
	SettleCoins int64   //结算区的金额

}

func NewDesk(roomNo room.Number, opts DeskOpts) *Desk {

	d := &Desk{
		round:           1,
		roomNo:          roomNo,
		deskOpt:         opts,
		players:         []*Player{},
		seatPlayers:     map[int32]*Player{},
		doingPlayers:    map[int32]*Player{},
		group:           nano.NewGroup(uuid.New()),
		InDeskPlayGroup: nano.NewGroup(uuid.New()),
		isFirstRound:    true,
		totalBet:        0,
		logger:          log.WithField("deskno", roomNo),
		//游戏部分
		gameState:     GameStateWaitJoin,
		gameStateTime: 0,
		TList:         []*Timer{},
		BankerId:      -1,
		FristOutId:    -1,
		OperateId:     -1,
		WildCard:      GCard{},
		ShowCard:      GCard{},
		PublicCard:    []GCard{},
		PointValue:    0,
		SettleCoins:   0,
	}
	//获取随机种子
	rand.Seed(time.Now().UnixNano())
	d.CardMgr.InitCards()
	d.CardMgr.Shuffle()
	go d.DoTimer()

	logger.Println("new desk:", roomNo.String())

	return d
}

// 玩家数量
func (d *Desk) totalPlayerCount() int {

	d.playMutex.Lock()
	defer d.playMutex.Unlock()

	return len(d.players)

}

//获取座位上玩家是否满了
func (this *Desk) GetSeatIsFull() bool {
	this.playMutex.Lock()
	defer this.playMutex.Unlock()
	return len(this.seatPlayers) >= DeskFullplayerNum
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

// 如果是重新进入 isReJoin: true
func (d *Desk) playerJoin(s *session.Session, isReJoin bool) error {

	uid := s.UID()

	var (
		p   *Player
		err error
	)

	if isReJoin {
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
			if !d.AddPlayerToDesk(p) {
				d.logger.Error("desk.playerJoin的玩家为空指针", p)
			}
			//d.roundStats[uid] = &history.Record{}
		}
	}
	return nil
}

//发送桌上玩家的状态信息给 每个人
func (d *Desk) PlayerJoinAfterInfo(p *Player) error {

	// d.playMutex.Lock()
	// defer d.playMutex.Unlock()
	//广播玩家加入信息
	d.group.Broadcast(NoticePlayerJoin, &protocol.EnterDeskInfo{
		SeatPos:  p.seatPos,
		Nickname: p.name,
		Sex:      p.sex,
		HeadUrl:  p.head,
		StarNum:  p.starNum,
		IsBanker: p.isBanker,
		Sitdown:  p.sitdown,
		Betting:  false, /*p.IsBetting()*/
		Show:     p.showed,
	})
	//给玩家发送桌子信息
	d.group.Add(p.session)

	deskInfo := protocol.DeskInfo{
		PointValue:    d.PointValue,
		DecksNum:      2,
		MaxWining:     DeskFullplayerNum * d.PointValue * DeskMaxLosePoint,
		Maxlosing:     d.PointValue * DeskMaxLosePoint,
		SettleCoins:   d.SettleCoins,
		GameState:     int32(d.gameState),
		GameStateTime: int32(d.GetTimerNum(d.gameState)),
		OperSeatId:    d.OperateId,
		BankerSeatId:  d.BankerId,
		FirstSeatId:   d.FristOutId,
		UserSeatId:    int32(p.seatPos),
	}
	for _, p := range d.seatPlayers {

		deskInfo.PlayersInfo = append(deskInfo.PlayersInfo, protocol.EnterDeskInfo{
			SeatPos:  p.seatPos,
			Nickname: p.name,
			Sex:      p.sex,
			HeadUrl:  p.head,
			StarNum:  p.starNum,
			IsBanker: p.isBanker,
			Sitdown:  p.sitdown,
			Betting:  false,
			Show:     p.showed,
		})
	}
	err := p.session.Response(&protocol.JoinDeskResponse{
		Success:  true,
		DeskInfo: deskInfo,
	})

	if d.gameState != GameStateWaitStart {
		return err
	}
	d.ClearTimer()
	d.AddTimer(GameStateWaitStart, GameStateWaitStartTime, d.start, nil)
	d.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(d.gameState),
		Time:      int32(d.GetTimerNum(d.gameState)),
	})

	return err

}

func (this *Desk) start(interface{}) {
	fmt.Println("游戏开始！\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\")
	//检测桌子开始条件是否符合
	if len(this.players) < DeskMinplayerNum {
		if len(this.players) == 0 {
			//解散桌子
		}
		this.gameState = GameStateWaitJoin
		return
	}
	//游戏开始
	this.ClearTimer()
	this.gameState = GameStateSendCard
	this.doingPlayers = this.seatPlayers
	this.round = 0
	//初始化玩家
	for _, v := range this.doingPlayers {
		v.InitPlayer()
	}
	//洗牌
	this.CardMgr.Shuffle()
	//发牌
	for _, p := range this.doingPlayers {
		p.HandCards = append([]GCard{}, this.CardMgr.SendCard(13)...)
		this.CardMgr.QuickSortCLV(&p.HandCards, 0, len(p.HandCards)-1)
		fmt.Println("玩家的手牌：", p.HandCards)
	}
	//定庄(有庄不变，没庄换庄)
	if this.BankerId == -1 {
		this.BankerId = int32(rand.Intn(len(this.doingPlayers)))
		//定首家(庄家的下家)
		this.OperateId = (this.BankerId + 1) % int32(len(this.doingPlayers))
	} else {
		//定首家(庄家的下家)
		for i := 1; i < DeskFullplayerNum; i++ {
			bankernext := (this.BankerId + int32(i)) % DeskFullplayerNum
			if this.doingPlayers[bankernext] != nil {
				this.OperateId = bankernext
				break
			}
		}
	}
	//广播开始通知
	this.group.Broadcast(NoticeGameStrat, &protocol.GGameStartNotice{
		BankerId: this.BankerId,
		FristID:  this.OperateId,
	})

	//定万能牌
	this.WildCard = this.CardMgr.SendCard(1)[0]
	//翻第一张牌
	this.PublicCard = this.CardMgr.SendCard(1)
	//发牌动画
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
	})
	for _, p := range this.doingPlayers {
		var handcard []int32
		for _, v := range p.HandCards {
			handcard = append(handcard, v.Card)
		}
		p.session.Push(NoticeSendCard, &protocol.GSendCardNotice{
			HandCards: handcard,
			WildCard:  this.WildCard.Card,
			FristCard: this.PublicCard[len(this.PublicCard)-1].Card,
		})
	}

	this.AddTimer(GameStateSendCard, GameStateSendCardTime, this.SendCardOutTime, nil)

}

//玩家操作通知
func (this *Desk) SendCardOutTime(interface{}) {
	if this.OperateId == this.FristOutId {
		this.round++
	}
	this.gameState = GameStatePlay
	this.AddTimer(GameStatePlay, GameStatePlayTime, this.PlayeroperOutTime, nil)
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
		OperateId: this.OperateId,
	})
	fmt.Println("当前轮到玩家：", this.OperateId)
}

//玩家操作超时
func (this *Desk) PlayeroperOutTime(interface{}) {
	fmt.Println("正在玩所有玩家", this.doingPlayers, "操作的玩家：", this.OperateId)
	p := this.doingPlayers[this.OperateId]
	handcard := p.HandCards
	notice := protocol.GPlayEndNotice{}
	if len(handcard) < 13 || len(handcard) > 14 {
		this.logger.Error("座位号:%d  手牌数不对，手牌是%v", this.OperateId, handcard)
		return
	}
	if p.Timeout >= 2 { //玩家超时弃牌
		this.GiveUp(p)
		return
	}
	if len(handcard) == 13 {
		//摸牌操作
		//this.OperCard(p, &protocol.GOperCardRequest{Opertion: 2})
		operCard := this.CardMgr.SendCard(1)[0]
		notice.Opertions = append(notice.Opertions, 2)
		notice.OperCard = append(notice.OperCard, operCard.Card)
		notice.PublicCard = append(notice.PublicCard, this.PublicCard[len(this.PublicCard)-1].Card)
		p.HandCards = append(p.HandCards, operCard)
	}
	//出牌操作
	//this.OperCard(p, &protocol.GOperCardRequest{
	//	Opertion: 3,
	//	OperCard: p.HandCards[len(p.HandCards)-1].Card,
	//})
	Card := p.HandCards[len(p.HandCards)-1].Card
	p.DelHandCard(Card)
	this.PublicCard = append(this.PublicCard, GCard{Poker.CardBase{Card: Card}})
	notice.Opertions = append(notice.Opertions, 3)
	notice.OperCard = append(notice.OperCard, Card)
	notice.PublicCard = append(notice.PublicCard, this.PublicCard[len(this.PublicCard)-1].Card)
	p.session.Push(NoticePlayEnd, &notice)

	p.Timeout++
	p.deposit = true

	//轮到下一位玩家
	for i := 1; i < DeskFullplayerNum; i++ {
		nextplayer := (this.OperateId + int32(i)) % DeskFullplayerNum
		if this.doingPlayers[nextplayer] != nil {
			this.OperateId = nextplayer
			break
		}
	}
	this.SendCardOutTime(nil)
}

//操作牌
func (this *Desk) OperCard(p *Player, msg *protocol.GOperCardRequest) (err error) {
	//游戏状态监测
	if this.gameState != GameStatePlay {
		return p.session.Response(&protocol.GOperCardResponse{
			Opertion: 0,
			Error:    "玩家不在游戏中！",
		})
	}
	//检测是否轮到该玩家
	if this.doingPlayers[this.OperateId] != p {
		return p.session.Response(&protocol.GOperCardResponse{
			Opertion: 0,
			Error:    "还未轮到该玩家操作！",
		})
	}
	//检测玩家是否show
	if this.ShowCard.Card != 0 {
		return p.session.Response(&protocol.GOperCardResponse{
			Opertion: 0,
			Error:    "玩家show时间，不允许操作牌！",
		})
	}
	//检测是否玩家牌数是否正确
	if len(p.HandCards) != 13 && len(p.HandCards) != 14 {
		this.logger.Debug("玩家%s手牌数%s不对!  ", p.name, len(p.HandCards), p.uid)
		return p.session.Response(&protocol.GOperCardResponse{
			Opertion: 0,
			Error:    "玩家手牌数不对！",
		})
	}
	fmt.Println("玩家：,操作：,手牌: ", p.name, msg.Opertion, msg.OperCard)
	//判断玩家的操作
	msgResponse := protocol.GOperCardResponse{
		Opertion: msg.Opertion,
		Error:    "玩家操作错误！",
	}
	msgNotice := protocol.GPlayerOperNotice{
		SeatId:   int32(p.seatPos),
		Opertion: msg.Opertion,
	}
	//摸牌公摊牌
	if msg.Opertion == 1 && len(p.HandCards) == 13 {
		operCard := this.PublicCard[len(this.PublicCard)-1]
		p.HandCards = append(p.HandCards, operCard)
		this.PublicCard = append([]GCard{}, this.PublicCard[:len(this.PublicCard)-1]...)
		publicCard := int32(0)
		if len(this.PublicCard) != 0 {
			publicCard = this.PublicCard[len(this.PublicCard)-1].Card
		}
		msgResponse.OperCard = operCard.Card
		msgResponse.PublicCard = publicCard
		msgResponse.Error = ""
		msgNotice.OperCard = operCard.Card
		msgNotice.PublicCard = publicCard
	}
	//摸牌堆的牌
	if msg.Opertion == 2 && len(p.HandCards) == 13 {
		operCard := this.CardMgr.SendCard(1)[0]
		p.HandCards = append(p.HandCards, operCard)
		msgResponse.OperCard = operCard.Card
		msgResponse.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
		msgResponse.Error = ""
		msgNotice.OperCard = 0
		msgNotice.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
	}
	//出牌
	if msg.Opertion == 3 && len(p.HandCards) == 14 {
		this.DelTimer(GameStatePlay)
		if !p.DelHandCard(msg.OperCard) {
			p.session.Response(&protocol.GOperCardResponse{
				Opertion: 0,
				Error:    "玩家操作错误！没有这张手牌！",
			})
		}
		this.PublicCard = append(this.PublicCard, GCard{Poker.CardBase{Card: msg.OperCard}})
		msgResponse.OperCard = msg.OperCard
		msgResponse.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
		msgResponse.Error = ""
		msgNotice.OperCard = msg.OperCard
		msgNotice.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
	}
	//show（胡了）
	if msg.Opertion == 4 && len(p.HandCards) == 14 {
		this.DelTimer(GameStatePlay)
		msgResponse.OperCard = msg.OperCard
		msgResponse.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
		msgResponse.Error = ""
		msgNotice.OperCard = msg.OperCard
		msgNotice.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
		this.ShowCard = GCard{Poker.CardBase{Card: msg.OperCard}}
	}
	if msgResponse.Error != "" {
		msgResponse.Opertion = 0
	} else {
		p.Timeout = 0
	}
	err = p.session.Response(&msgResponse)
	if msgResponse.Error == "" {
		this.group.Broadcast(NoticeGameOperCard, &msgNotice)
	}
	//如果是出牌，切换下一位玩家
	if msgResponse.Opertion == 3 {
		//轮到下一位玩家
		for i := 1; i < DeskFullplayerNum; i++ {
			nextplayer := (this.OperateId + int32(i)) % DeskFullplayerNum
			if this.doingPlayers[nextplayer] != nil {
				this.OperateId = nextplayer
				break
			}
		}
		this.SendCardOutTime(nil)
	}
	return
}

//玩家show操作
func (this *Desk) ShowCards(p *Player, msg *protocol.GShowCardsRequest) error {
	//检测是否轮到这位玩家
	if int32(p.seatPos) != this.OperateId || this.ShowCard.Card == 0 {
		return p.session.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: "玩家没有Show Card！",
		})
	}
	//检测玩家组合的牌组中是否超过5组
	if len(msg.ShowCards) >= 5 {
		return p.session.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: "玩家的牌组不小于5组",
		})
	}
	//取出牌组
	var showCards [][]GCard
	var tcards []GCard
	for _, v := range msg.ShowCards {
		var show []GCard
		for _, v1 := range v {
			show = append(show, GCard{Poker.CardBase{Card: v1}})
		}
		showCards = append(showCards, show)
		tcards = append(tcards, show...)
	}
	//检测是否是13张
	if len(msg.ShowCards) != 13 {
		return p.session.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: "玩家Show Cards的牌数不对！",
		})
	}
	//检验玩家Show的牌是否都是玩家的手牌
	this.CardMgr.QuickSortCLV(&tcards, 0, len(tcards)-1)
	this.CardMgr.QuickSortCLV(&p.HandCards, 0, len(msg.ShowCards)-1)
	for k, v := range p.HandCards {
		if v != tcards[k] {
			return p.session.Response(&protocol.GShowCardsResponse{
				IsWin: false,
				Error: "玩家Show Cards跟玩家手牌不匹配！",
			})
		}
	}
	//检测1st life
	Is1stlife := false
	for k, v := range showCards {
		if this.CardMgr.Is1stLife(v) {
			showCards = append(showCards[:k], showCards[k+1:]...)
			Is1stlife = true
			break
		}
	}
	if !Is1stlife {
		return p.session.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: "玩家没有胡牌！",
		})
	}
	//检测2st life
	Is2stlife := false
	for k, v := range showCards {
		if this.CardMgr.Is2stLife(v, this.WildCard) {
			showCards = append(showCards[:k], showCards[k+1:]...)
			Is2stlife = true
			break
		}
	}
	if !Is2stlife {
		return p.session.Response(&protocol.GShowCardsResponse{
			IsWin: false,
			Error: "玩家没有胡牌！",
		})
	}
	//检测其他牌组是否成型
	for _, v := range showCards {
		if !this.CardMgr.Is2stLife(v, this.WildCard) && !this.CardMgr.IsSetLife(v, this.WildCard) {
			return p.session.Response(&protocol.GShowCardsResponse{
				IsWin: false,
				Error: "玩家没有胡牌！",
			})
		}
	}
	//应答成功
	err := p.session.Response(&protocol.GShowCardsResponse{
		IsWin: true,
	})
	//广播
	this.group.Broadcast(NoticeGameWin, &protocol.GPlayerWinNotice{
		SeatId: int32(p.seatPos),
	})
	//添加结算时间
	this.AddTimer(GameStateStettle, GameStateStettleTime-1, this.SettleTimeOut, nil)
	this.gameState = GameStateStettle
	return err
}

//检测桌子的玩家是否可以结算了
func (d *Desk) deskIsAccount() bool {

	return false

}

func (d *Desk) playerWithId(uid int64) (*Player, error) {

	for _, p := range d.players {
		if p.Uid() == uid {
			return p, nil
		}
	}

	return nil, errutil.ErrPlayerNotFound
}

//玩家是否为空了
func (d *Desk) PlayersIsEmpty() bool {

	bempty := false

	if d.totalPlayerCount() == 0 {

		bempty = true
	}

	return bempty
}

// 摧毁桌子
func (d *Desk) destroy() {

	//删除桌子
	//scheduler.PushTask(func() {
	//	defaultDeskManager.setDesk(d.roomNo, nil)
	//})

}

func (d *Desk) onPlayerExit(s *session.Session, isDisconnect bool) {

	d.logger.Println("玩家下线了：uid", s.UID())

	uid := s.UID()
	d.group.Leave(s)

	if isDisconnect {
		//	d.dissolve.updateOnlineStatus(uid, false)
	} else {
		var restPlayers []*Player
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

//设置游戏状态
func (this *Desk) SetGameState(gamestate int) bool {
	//状态不变
	if gamestate == this.gameState {
		return false
	}
	this.gameState = gamestate
	return true
}

//添加玩家到桌子里
func (this *Desk) AddPlayerToDesk(p *Player) bool {
	if p == nil {
		return false
	}
	this.players = append(this.players, p)
	//设置座位号
	for i := int32(0); i < DeskFullplayerNum; i++ {
		if this.seatPlayers[i] == nil {
			this.seatPlayers[i] = p
			p.SetSeatPos(int(i))
			break
		}
		//不在座位上设为-1
		p.SetSeatPos(-1)
	}
	for i, p := range this.players {
		p.setDesk(this, i)
	}
	//如果玩家玩家人数够了
	if len(this.seatPlayers) >= 2 && this.gameState == GameStateWaitJoin {
		this.gameState = GameStateWaitStart
	}
	return true
}

//结算超时
func (this *Desk) SettleTimeOut(interface{}) {
	//检测玩家是否都结算了
	for _, v := range this.doingPlayers {
		if int32(v.seatPos) == this.OperateId {
			continue
		}
		this.GiveUp(v)
	}
}

//结算
func (this *Desk) Settle(p *Player) bool {
	if p == nil {
		this.logger.Debug("Desk.Settle: *Player is nil!")
		return false
	}
	if this.gameState != GameStateStettle {

	}
	return true
}

func (this *Desk) GiveUp(p *Player) bool {
	p.sitdown = false
	//还没扣钱
	if this.round <= 1 {
		this.totalBet += Desk1RoundLosePoint * this.PointValue
	} else if this.round == 2 {
		this.totalBet += Desk2RoundLosePoint * this.PointValue
	} else {
		this.totalBet += DeskMaxLosePoint * this.PointValue
	}
	//如果是轮到自己操作的时候放弃，那么切换下一个玩家
	if this.OperateId == int32(p.seatPos) {
		//轮到下一位玩家
		for i := 1; i < DeskFullplayerNum; i++ {
			nextplayer := (this.OperateId + int32(i)) % DeskFullplayerNum
			if this.doingPlayers[nextplayer] != nil {
				this.OperateId = nextplayer
				break
			}
		}
	}
	//
	delete(this.doingPlayers, int32(p.seatPos))
	delete(this.seatPlayers, int32(p.seatPos))
	//广播
	//如果只剩一家就结算
	if len(this.doingPlayers) == 1 {
		this.Settle(this.doingPlayers[0])
	}
	return true
}
