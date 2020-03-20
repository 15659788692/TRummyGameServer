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
	GameStateEnd       = 5
)
const ( //状态时间
	GameStateWaitStartTime = 7  //游戏开始倒计时
	GameStateSendCardTime  = 9  //发牌
	GameStatePlayTime      = 30 //每个玩家操作时间
	GameStateStettleTime   = 20 //结算时间
	GameStateEndTime       = 3
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

	playMutex    sync.RWMutex
	players      []*Player         //房间内所有玩家
	seatPlayers  map[int32]*Player //座位上的玩家
	doingPlayers map[int32]*Player //正在玩的玩家
	group        *nano.Group       //房间内组播通道
	isFirstRound bool              //是否是本桌的第一局牌
	logger       *log.Entry

	//游戏部分
	CardMgr       GMgrCard //卡牌管理器
	gameState     int      //游戏状态
	gameStateTime int32    //状态时间
	TList         []*Timer // 定时器列表

	BankerId    int32             //庄家
	FristOutId  int32             //首出玩家
	WildCard    GCard             //万能牌(除大小王外的另一张万能牌)
	ShowCard    GCard             //show的牌
	ShowPlayer  int32             //show的玩家
	PublicCard  []GCard           //公摊牌堆
	OperateId   int32             //当前正在操作的玩家
	PointValue  int64             //底注
	SettleCoins int64             //结算区的金额
	GameRecord  protocol.GEndForm //游戏记录
}

func NewDesk(roomNo room.Number, opts DeskOpts) *Desk {

	d := &Desk{
		round:        0,
		roomNo:       roomNo,
		deskOpt:      opts,
		players:      []*Player{},
		seatPlayers:  map[int32]*Player{},
		doingPlayers: map[int32]*Player{},
		group:        nano.NewGroup(uuid.New()),
		isFirstRound: true,
		logger:       log.WithField("deskno", roomNo),
		//游戏部分
		gameState:     GameStateWaitJoin,
		gameStateTime: 0,
		TList:         []*Timer{},
		BankerId:      -1,
		FristOutId:    -1,
		OperateId:     -1,
		ShowPlayer:    -1,
		WildCard:      GCard{},
		ShowCard:      GCard{},
		PublicCard:    []GCard{},
		PointValue:    100,
		SettleCoins:   0,
	}
	//获取随机种子
	rand.Seed(time.Now().UnixNano())
	d.CardMgr.InitCards()
	d.CardMgr.Shuffle()
	d.GameRecord = protocol.GEndForm{}
	go d.DoTimer()

	logger.Println("new desk:", roomNo.String())

	return d
}
func (this *Desk) InitDesk() {
	this.round = 0
	this.doingPlayers = map[int32]*Player{}
	this.CardMgr.Shuffle()
	this.ClearTimer()
	this.WildCard = GCard{}
	this.ShowCard = GCard{}
	this.PublicCard = []GCard{}
	this.OperateId = -1
	this.ShowPlayer = -1
	this.GameRecord = protocol.GEndForm{}
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
		SeatPos:  int(p.seatPos),
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
		WildCard:      d.WildCard.Card,
		ShowCard:      d.ShowCard.Card,
		CardsNum:      int32(d.CardMgr.GetLeftCardCount()),
		OperSeatId:    d.OperateId,
		BankerSeatId:  d.BankerId,
		FirstSeatId:   d.FristOutId,
		UserSeatId:    p.seatPos,
	}
	for _, v := range d.seatPlayers {

		deskInfo.PlayersInfo = append(deskInfo.PlayersInfo, protocol.EnterDeskInfo{
			SeatPos:  int(v.seatPos),
			Nickname: v.name,
			Sex:      v.sex,
			HeadUrl:  v.head,
			StarNum:  v.starNum,
			IsBanker: v.isBanker,
			Sitdown:  v.sitdown,
			Betting:  false,
			Show:     v.showed,
			Coins:    v.Coins,
		})
	}
	//游戏开始，自己不在游戏中
	if d.gameState > GameStateWaitStart {
		deskInfo.PublicCard = d.PublicCard[len(d.PublicCard)-1].Card
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
	this.InitDesk()
	this.gameState = GameStateSendCard
	for k, v := range this.seatPlayers {
		this.doingPlayers[k] = v
	}
	//初始化玩家
	for _, v := range this.doingPlayers {
		v.InitPlayer()
	}
	//洗牌
	this.CardMgr.Shuffle()
	//发牌
	for _, p := range this.doingPlayers {
		p.HandCards = append([]GCard{}, this.CardMgr.SendCard(13)...)
		this.CardMgr.QuickSortCLV(p.HandCards)
		//拆成牌组
		t := 0
		tcard := p.HandCards[0]
		for k, v := range p.HandCards {
			if v.GetCardColor() != tcard.GetCardColor() {
				p.CardsSet = append(p.CardsSet, this.GCardToInt32(p.HandCards[t:k]))
				t = k
				tcard = v
			}
		}
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
	this.ClearTimer()
	this.AddTimer(GameStateSendCard, GameStateSendCardTime, this.OpertionNotice, nil)

}

//玩家操作通知
func (this *Desk) OpertionNotice(interface{}) {
	if this.OperateId == this.FristOutId {
		this.round++
	}
	if this.CardMgr.GetLeftCardCount() <= 0 {
		//广播游戏结束
		this.gameState = GameStateEnd
		this.group.Broadcast(NoticeEndInfo, &protocol.GEndForm{})
		this.ClearTimer()
		this.AddTimer(GameStateEnd, GameStateEndTime, this.GameEnd, nil)
		this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
			GameState: int32(this.gameState),
			Time:      int32(this.GetTimerNum(this.gameState)),
		})

		return
	}
	this.gameState = GameStatePlay
	this.ClearTimer()
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
	if p == nil {
		return
	}
	handcard := p.HandCards
	if len(handcard) < 13 || len(handcard) > 14 {
		this.logger.Error("座位号:%d  手牌数不对，手牌是%v", this.OperateId, handcard)
		return
	}
	if p.Timeout >= 2 || this.ShowPlayer != -1 { //玩家超时弃牌
		fmt.Println("玩家操作超时弃牌！")
		_ = this.GiveUp(p, true, &protocol.GGiveUpRequect{
			Cards: p.CardsSet,
		})
		this.ShowCard = GCard{}
		this.ShowPlayer = -1
		return
	}
	if len(handcard) == 13 {
		//摸牌操作
		this.OperCard(p, &protocol.GOperCardRequest{Opertion: 2}, true)
		this.ClearTimer()
		this.AddTimer(GameStatePlay, 3, this.PlayeroperOutTime, nil)
	}
	//出牌操作
	if len(handcard) == 14 {
		this.OperCard(p, &protocol.GOperCardRequest{
			Opertion: 3,
			OperCard: p.HandCards[len(p.HandCards)-1].Card,
		}, true)
		p.Timeout++
		p.deposit = true
	}

}

//操作牌
func (this *Desk) OperCard(p *Player, msg *protocol.GOperCardRequest, IsOuttime bool) (err error) {
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
	if this.ShowPlayer != -1 {
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
	//没牌可摸了
	if this.CardMgr.GetLeftCardCount() <= 0 && msg.Opertion == 2 && len(p.HandCards) == 13 {
		return p.session.Response(&protocol.GOperCardResponse{
			Opertion: 0,
			Error:    "操作错误！",
		})
	}
	fmt.Println("玩家：,操作：,手牌: ", p.name, msg.Opertion, msg.OperCard)
	//判断玩家的操作
	msgResponse := protocol.GOperCardResponse{
		Opertion: msg.Opertion,
		Error:    "玩家操作错误！",
	}
	msgNotice := protocol.GPlayerOperNotice{
		SeatId:   p.seatPos,
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
		operCard := this.CardMgr.SendCard(1)
		p.HandCards = append(p.HandCards, operCard[0])
		msgResponse.OperCard = operCard[0].Card
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
		if !p.DelHandCard(msg.OperCard) {
			p.session.Response(&protocol.GOperCardResponse{
				Opertion: 0,
				Error:    "玩家操作错误！没有这张手牌！",
			})
		}
		msgResponse.OperCard = msg.OperCard
		msgResponse.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
		msgResponse.Error = ""
		msgNotice.OperCard = msg.OperCard
		msgNotice.PublicCard = this.PublicCard[len(this.PublicCard)-1].Card
		this.ShowCard = GCard{Poker.CardBase{Card: msg.OperCard}}
		this.ShowPlayer = p.seatPos
	}
	msgResponse.CardsNum = int32(this.CardMgr.GetLeftCardCount())
	msgNotice.CardsNum = int32(this.CardMgr.GetLeftCardCount())
	if msgResponse.Error != "" {
		msgResponse.Opertion = 0
	} else {
		p.Timeout = 0
	}
	if IsOuttime {
		err = p.session.Push(NoticeOperOutTime, &msgResponse)
	} else {
		err = p.session.Response(&msgResponse)
	}
	if msgResponse.Error == "" {
		_ = this.group.Multicast(NoticeGameOperCard, &msgNotice, func(s *session.Session) bool {
			if s == p.session {
				return false
			}
			return true
		})
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
		this.ClearTimer()
		this.AddTimer(GameStatePlay, 2, this.OpertionNotice, nil)
	}
	return
}

//玩家整理牌
func (this *Desk) ShowCards(p *Player, msg *protocol.GSetHandCardRequest) error {
	//检验游戏状态
	if this.gameState < GameStatePlay {
		return p.session.Response(&protocol.GSetHandCardResponse{
			Success: false,
			Error:   "还没开始游戏！",
		})
	}
	//检测玩家组合的牌组中是否超过5组
	if len(msg.CardsSets) > 5 {
		return p.session.Response(&protocol.GSetHandCardResponse{
			Success: false,
			Error:   "玩家的牌组不小于5组",
		})
	}
	//检测是否是自己的手牌
	for _, v := range msg.CardsSets {
		if !p.IsMyHandCard(v.Cards) {
			return p.session.Response(&protocol.GSetHandCardResponse{
				Success: false,
				Error:   "组牌错误！",
			})
		}
	}
	//处理发过来的牌组类型
	msgResponse := protocol.GSetHandCardResponse{}
	var tset []protocol.CardsSet
	tset, msgResponse.TotalPoint = this.HandleCardsSet(msg.CardsSets)
	p.CardsSet = this.CardsSetToInt32(tset)
	msgResponse.CardsSets = tset
	msgResponse.Success = true
	//if msg.IsFinish && msgResponse.TotalPoint == 0 {
	msgResponse.IsHu = true
	//}
	err := p.session.Response(&msgResponse)
	//判断是否胡牌
	if !msg.IsFinish {
		return err
	}
	//if msgResponse.TotalPoint != 0 {
	//	fmt.Println("玩家show牌失败弃牌！")
	//	this.ClearTimer()
	//	_ = this.GiveUp(p, true, &protocol.GGiveUpRequect{Cards: p.CardsSet})
	//	this.ShowCard = GCard{}
	//	this.ShowPlayer = -1
	//	return err
	//}
	//广播
	this.group.Broadcast(NoticeGameWin, &protocol.GPlayerWinNotice{
		SeatId: p.seatPos,
	})
	//添加结算时间
	this.ClearTimer()
	this.AddTimer(GameStateStettle, GameStateStettleTime-1, this.SettleTimeOut, nil)
	this.gameState = GameStateStettle
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
	})
	return err
}

//结算超时
func (this *Desk) SettleTimeOut(interface{}) {
	//检测玩家是否都结算了
	for _, v := range this.doingPlayers {
		if v.seatPos == this.OperateId || v.settle {
			continue
		}
		_ = this.Settle(v, &protocol.GSettleRequect{
			Cards: v.CardsSet,
		}, true)
	}
}

//结算
func (this *Desk) Settle(p *Player, msg *protocol.GSettleRequect, IsOutTime bool) (err error) {
	//检测
	if p == nil {
		this.logger.Debug("Desk.Settle: *Player is nil!")
		return
	}
	if this.gameState != GameStateStettle {
		err = p.session.Response(&protocol.GSettleResponse{
			WinCoins: 0,
			Error:    "状态错误！",
		})
		return
	}
	//如果是赢的玩家就不用结算
	if this.OperateId == p.seatPos {
		err = p.session.Response(&protocol.GSettleResponse{
			WinCoins: 0,
			Error:    "消息错误！",
		})
		return
	}
	if p.settle {
		err = p.session.Response(&protocol.GSettleResponse{
			WinCoins: 0,
			Error:    "不能重复结算！",
		})
		return
	}

	//取出牌组
	var tcards [][]GCard
	var show []GCard
	for _, v := range msg.Cards {
		for _, v1 := range v {
			show = append(show, GCard{Poker.CardBase{Card: v1}})
		}
		tcards = append(tcards, show)
	}
	//检验发过来的牌组是否都是玩家的手牌
	fmt.Println("玩家的手牌：", p.HandCards, show)
	this.CardMgr.QuickSortCLV(show)
	this.CardMgr.QuickSortCLV(p.HandCards)
	fmt.Println("玩家的手牌：", p.HandCards, show)
	for k, v := range p.HandCards {
		if show[k] != v {
			err = p.session.Response(&protocol.GSettleResponse{
				WinCoins: 0,
				Error:    "发过来的牌组错误！",
			})
			return
		}
	}
	//
	nextplayerId := int32(0)
	for i := 1; i < DeskFullplayerNum; i++ {
		nextplayerId = (this.OperateId + int32(i)) % DeskFullplayerNum
		if this.doingPlayers[nextplayerId] != nil {
			break
		}
	}
	if this.round <= 1 && (this.OperateId == this.FristOutId || this.OperateId == nextplayerId) { //天胡或者地胡
		p.Point = 80
		p.win = -1 * 80 * this.PointValue
	} else {
		_, v := this.CardMgr.CheckoutHu(tcards, this.WildCard)
		p.Point = int32(v)
		p.win = -1 * v * this.PointValue
	}
	p.Coins += p.win
	this.SettleCoins -= p.win
	p.settle = true
	//添加记录
	this.GameRecord.EndInfo = append(this.GameRecord.EndInfo, protocol.PlayerEndInfo{
		Name:      p.name,
		Head:      p.head,
		CardsSets: p.CardsSet,
		Point:     p.Point,
		Coins:     p.win,
	})
	//结算通知
	if IsOutTime {
		p.session.Push("", &protocol.GSettleResponse{
			WinCoins: p.win,
		})
	} else {
		err = p.session.Response(&protocol.GSettleResponse{
			WinCoins: p.win,
		})
	}

	//判断是否都结算完了
	for _, v := range this.doingPlayers {
		if v.seatPos == this.OperateId || v.settle {
			continue
		}
		return
	}
	winPlayer := this.doingPlayers[this.OperateId]
	winPlayer.win = this.SettleCoins
	winPlayer.Coins += this.SettleCoins
	//广播游戏结束
	//添加赢家的记录
	this.GameRecord.EndInfo = append([]protocol.PlayerEndInfo{{
		Name:      winPlayer.name,
		Head:      winPlayer.head,
		CardsSets: winPlayer.CardsSet,
		Point:     0,
		Coins:     winPlayer.win,
	}}, this.GameRecord.EndInfo...)
	this.group.Broadcast(NoticeEndInfo, &this.GameRecord)
	this.gameState = GameStateEnd
	this.ClearTimer()
	this.AddTimer(GameStateEnd, GameStateEndTime, this.GameEnd, nil)
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
	})
	return
}

//放弃
func (this *Desk) GiveUp(p *Player, IsOutTime bool, msg *protocol.GGiveUpRequect) error {
	//检测牌是否是自己的手牌
	p.CardsSet = msg.Cards
	//扣钱
	if IsOutTime && this.ShowPlayer != -1 {
		p.Point = DeskMaxLosePoint
	} else if this.round <= 1 {
		p.Point = Desk1RoundLosePoint
	} else if this.round == 2 {
		p.Point = Desk2RoundLosePoint
	} else {
		p.Point = DeskMaxLosePoint
	}
	coins := int64(p.Point) * this.PointValue
	this.SettleCoins += coins
	p.win = -1 * coins
	p.Coins -= coins
	//添加记录
	this.GameRecord.EndInfo = append(this.GameRecord.EndInfo, protocol.PlayerEndInfo{
		Name:      p.name,
		Head:      p.head,
		CardsSets: p.CardsSet,
		Point:     p.Point,
		Coins:     p.win,
	})
	//如果是轮到自己操作的时候放弃，那么切换下一个玩家
	if this.OperateId == p.seatPos {
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
	msgNotice := protocol.GGiveUpNotice{}
	msgNotice.IsShow = IsOutTime
	msgNotice.LosingCoins = coins
	msgNotice.SeatId = p.seatPos
	msgNotice.SettleCoins = this.SettleCoins
	var err error
	if IsOutTime {
		p.session.Push(NoticeLoseGame, &protocol.GGiveUpResponse{
			Success:     true,
			Coins:       coins,
			PlayerCoins: p.Coins,
			TotalCoins:  this.SettleCoins,
		})
	} else {
		err = p.session.Response(&protocol.GGiveUpResponse{
			Success:     true,
			Coins:       coins,
			PlayerCoins: p.Coins,
			TotalCoins:  this.SettleCoins,
		})
	}
	//广播玩家弃牌
	fmt.Println("广播玩家弃牌!!!!!!!!!!")
	_ = this.group.Multicast(NoticeGiveUp, &msgNotice, func(s *session.Session) bool {
		//if s == p.session && IsOutTime {
		//	return false
		//}
		return true
	})

	//如果只剩一家就结算
	delete(this.doingPlayers, p.seatPos)
	delete(this.seatPlayers, p.seatPos)
	if len(this.doingPlayers) == 1 {
		//广播
		fmt.Println("玩家不战而胜！！！！！")
		this.group.Broadcast(NoticeGameWin, &protocol.GPlayerWinNotice{
			SeatId: this.doingPlayers[0].seatPos,
		})
		this.gameState = GameStateEnd
		this.AddTimer(GameStateEnd, GameStateEndTime, this.GameEnd, nil)
		this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
			GameState: int32(this.gameState),
			Time:      int32(this.GetTimerNum(this.gameState)),
		})
		return err
	}
	this.OpertionNotice(nil)
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
			p.SetSeatPos(i)
			break
		}
		//不在座位上设为-1
		p.SetSeatPos(-1)
	}
	for i, p := range this.players {
		p.setDesk(this, int32(i))
	}
	//如果玩家玩家人数够了
	if len(this.seatPlayers) >= 2 && this.gameState == GameStateWaitJoin {
		this.gameState = GameStateWaitStart
	}
	return true
}

//游戏结束
func (this *Desk) GameEnd(interface{}) {

	//判断是否人数足够重开一把
	if len(this.seatPlayers) < DeskMinplayerNum {
		this.gameState = GameStateWaitJoin
		return
	}
	//游戏重开一局
	this.gameState = GameStateWaitStart
	this.ClearTimer()
	this.AddTimer(GameStateWaitStart, GameStateWaitStartTime, this.start, nil)
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
	})
}

func (this *Desk) Int32ToGCard(str []int32) []GCard {
	var tset []GCard
	for _, v := range str {
		tset = append(tset, GCard{Poker.CardBase{Card: v}})
	}
	return tset
}

func (this *Desk) GCardToInt32(str []GCard) []int32 {
	var tset []int32
	for _, v := range str {
		tset = append(tset, v.Card)
	}
	return tset
}

func (this *Desk) CardsSetToInt32(cardsSet []protocol.CardsSet) [][]int32 {
	var str [][]int32
	for _, v := range cardsSet {
		str = append(str, v.Cards)
	}
	return str
}

//整理牌组的类型
func (this *Desk) HandleCardsSet(CardsSets []protocol.CardsSet) ([]protocol.CardsSet, int32) {
	//计算这些牌的组合情况
	totalPoint := int32(0)
	sets := CardsSets
	have1st := -1
	have2st := -1
	var firstK []int
	var secondK []int
	var GroupK []int
	//找出第一生命和第二生命
	for k, v := range CardsSets {
		if this.CardMgr.Is1stLife(this.Int32ToGCard(v.Cards)) {
			sets[k].Type = Type1stLife
			firstK = append(firstK, k)
		} else if this.CardMgr.Is2stLife(this.Int32ToGCard(v.Cards), this.WildCard) {
			sets[k].Type = Type2stLife
			secondK = append(secondK, k)
		} else if this.CardMgr.IsSetLife(this.Int32ToGCard(v.Cards), this.WildCard) {
			sets[k].Type = TypeGroup
			GroupK = append(GroupK, k)
		} else {
			sets[k].Type = TypeOther
			sets[k].Point = int32(this.CardMgr.ComputePoint([][]GCard{this.Int32ToGCard(v.Cards)}, this.WildCard))
			totalPoint += sets[k].Point
		}
		if v.Type == Type1stLife && sets[k].Type >= Type1stLife {
			have1st = k
		}
		if v.Type == Type2stLife && sets[k].Type >= Type2stLife {
			have2st = k
		}
	}
	//确定唯一第一生命
	for _, v := range firstK {
		if have1st == -1 {
			have1st = v
			continue
		}
		sets[v].Type = Type2stLife
		secondK = append(secondK, v)
	}
	//确定唯一第二生命
	for _, v := range secondK {
		if have1st == -1 {
			sets[v].Type = TypeNeed1stLife
			continue
		}
		if have2st == -1 {
			have2st = v
			continue
		}
		sets[v].Type = TypeGroup
		GroupK = append(GroupK, v)
	}
	//找出组
	for _, v := range GroupK {
		if have2st == -1 {
			sets[v].Type = TypeNeed2stLife
			continue
		}
		sets[v].Type = TypeGroup
	}
	return sets, totalPoint
}

//出牌记录
func (this *Desk) OutCardRecord(p *Player) error {
	//验证玩家是否在游戏中
	if this.doingPlayers[p.seatPos] == nil {
		return p.session.Response(&protocol.GOutCardRecordResponse{
			Success: false,
		})
	}
	var cards = [][]int32{{}, {}, {}, {}}
	//排序
	tGCard := append([]GCard{}, this.PublicCard...)
	this.CardMgr.QuickSortCLV(tGCard)
	//花色区分
	for _, v := range this.GCardToInt32(tGCard) {
		if v >= Poker.Card_Fang_1 && v <= Poker.Card_Fang_K {
			cards[0] = append(cards[0], v)
			continue
		}
		if v >= Poker.Card_Mei_1 && v <= Poker.Card_Mei_K {
			cards[1] = append(cards[1], v)
			continue
		}
		if v >= Poker.Card_Hong_1 && v <= Poker.Card_Hong_K {
			cards[2] = append(cards[2], v)
			continue
		}
		if v >= Poker.Card_Hei_1 && v <= Poker.Card_Hei_K {
			cards[3] = append(cards[3], v)
			continue
		}
	}

	return p.session.Response(&protocol.GOutCardRecordResponse{
		Success:    true,
		CardRecord: cards,
	})

}
