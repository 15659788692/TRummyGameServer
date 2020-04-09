package game

import "TeenPatti/TRummyGameServer/conf"

import (
	"TeenPatti/TRummyGameServer/Poker"
	"TeenPatti/TRummyGameServer/pkg/errutil"
	"math/rand"
	"sync"

	// "sync/atomic"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/session"
	"github.com/pborman/uuid"

	// "TeenPatti/TRummyGameServer/pkg/constant"
	"TeenPatti/TRummyGameServer/pkg/room"
	"TeenPatti/TRummyGameServer/protocol"
)

//游戏状态
const (
	GameStateWaitJoin  = 0
	GameStateWaitStart = 1
	GameStateSendCard  = 2
	GameStatePlay      = 3 //玩家操作阶段
	GameStatePlayAV    = 7
	GameStateStettle   = 4
	GameStateStettleAV = 8
	GameStateEnd       = 5
	GameStateAbort     = 6
)

//游戏状态时间
const (
	GameStatePlayAVTime    = 2
	GameStateStettleAVTime = 3
)

//组牌操作
const (
	Unspecified = 0
	beforeShow  = 1 // show操作前尝试检测玩家手上剩余的牌能不能胡用的
	Show        = 2
	Final       = 3
)

type Desk struct {
	roomNo room.Number //房间号
	deskID int64       //desk表的pk
	//	opts      *protocol.DeskOptions // 房间选项
	round     uint32 // 第n轮
	creator   int64  // 创建玩家UID
	createdAt int64  // 创建时间

	Mutex        sync.RWMutex
	players      []*Player         //房间内所有玩家
	seatPlayers  map[int32]*Player //座位上的玩家
	doingPlayers map[int32]*Player //正在玩的玩家
	group        *nano.Group       //房间内组播通道
	isFirstRound bool              //是否是本桌的第一局牌

	//游戏部分
	CardMgr       GMgrCard //卡牌管理器
	gameState     int      //游戏状态
	gameStateTime int32    //状态时间
	TList         []*Timer // 定时器列表

	BankerId    int32             //庄家
	KingId      int32             //房主
	FristOutId  int32             //首出玩家
	WildCard    GCard             //万能牌(除大小王外的另一张万能牌)
	ShowCard    GCard             //show的牌
	ShowPlayer  int32             //show的玩家
	WinPlayerId int32             //赢的玩家的座位号
	PublicCard  []GCard           //公摊牌堆
	OperateId   int32             //当前正在操作的玩家
	PointValue  int64             //底注
	SettleCoins int64             //结算区的金额
	GameRecord  protocol.GEndForm //游戏记录
	Config      conf.DeskData
}

func NewDesk(roomNo room.Number) *Desk {

	d := &Desk{
		round:        0,
		roomNo:       roomNo,
		players:      []*Player{},
		seatPlayers:  map[int32]*Player{},
		doingPlayers: map[int32]*Player{},
		group:        nano.NewGroup(uuid.New()),
		isFirstRound: true,
		//游戏部分
		gameState:     GameStateWaitJoin,
		gameStateTime: 0,
		TList:         []*Timer{},
		BankerId:      -1,
		FristOutId:    -1,
		KingId:        -1,
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
	d.Config = (*conf.Conf).Desk
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
	this.SettleCoins = 0
	this.GameRecord = protocol.GEndForm{}
}

// 玩家数量
func (d *Desk) totalPlayerCount() int {

	return len(d.players)

}

//检测桌子人数是否满了
func (d *Desk) IsFullPlayer() bool {

	if len(d.players) == d.Config.FullPlayerNum {
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
		log.Println("玩家%v重新加入房间！", p.name)
		if err != nil {
			log.Errorf("玩家: %d重新加入房间, 但是没有找到玩家在房间中的数据", uid)
			return err
		}
		// 加入分组
		d.group.Add(s)

	} else {
		exists := false

		for _, p := range d.players {

			if p.Uid() == uid {
				exists = true
				log.Warn("玩家已经在房间中")
				break
			}
		}

		if !exists {
			p = s.Value(kCurPlayer).(*Player)
			if !d.AddPlayerToDesk(p) {
				log.Error("desk.playerJoin的玩家为空指针", p)
			}
			//d.roundStats[uid] = &history.Record{}
		}
	}
	return nil
}

//发送桌上玩家的状态信息给 每个人
func (d *Desk) PlayerJoinAfterInfo(p *Player, isReJoin bool) error {

	// d.playMutex.Lock()
	// defer d.playMutex.Unlock()
	var playerState int32 = PlayerStateJoin
	if isReJoin {
		playerState = PlayerStateReJoin
	}
	//广播玩家加入信息
	d.group.Multicast(NoticePlayerState, &protocol.PlayerStateNotice{
		SeatId:      p.seatPos,
		PlayerState: playerState,
		PlayerInfo: protocol.EnterDeskInfo{
			SeatPos:  p.seatPos,
			Nickname: p.name,
			Sex:      p.sex,
			HeadUrl:  p.head,
			StarNum:  p.starNum,
			IsBanker: p.isBanker,
			Sitdown:  p.sitdown,
			LiXian:   p.disconnect,
			Show:     p.showed,
		},
	}, func(s *session.Session) bool {
		if s == p.session {
			return false
		}
		return true
	})
	//给玩家发送桌子信息
	d.group.Add(p.session)
	deskInfo := d.GetDeskInfo(p)
	if (!p.isJoin && d.gameState > GameStateWaitStart) || p.seatPos < 0 {
		deskInfo.GameState *= -1
	}

	err := p.session.Response(&protocol.JoinDeskResponse{
		Success:  true,
		DeskInfo: deskInfo,
	})
	if len(d.seatPlayers) >= d.Config.MinPlayerNum && d.gameState == GameStateWaitJoin {
		d.gameState = GameStateWaitStart
	}
	if d.gameState != GameStateWaitStart || d.GetTimerNum(GameStateWaitStart) > 0 {
		return err
	}
	d.ClearTimer()
	d.AddTimer(GameStateWaitStart, d.Config.GameStateTime[GameStateWaitStart], d.start, nil)
	d.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(d.gameState),
		Time:      int32(d.GetTimerNum(d.gameState)),
		TotalTime: int32(d.Config.GameStateTime[d.gameState]),
	})

	return err

}

func (d *Desk) GetDeskInfo(p *Player) protocol.DeskInfo {
	deskInfo := protocol.DeskInfo{
		PointValue:     d.PointValue,
		DecksNum:       2,
		MaxWining:      int64(d.Config.FullPlayerNum) * d.PointValue * int64(d.Config.MaxLosePoint),
		Maxlosing:      d.PointValue * int64(d.Config.MaxLosePoint),
		SettleCoins:    d.SettleCoins,
		GameState:      int32(d.gameState),
		GameStateTime:  int32(d.GetTimerNum(d.gameState)),
		TotalStateTime: int32(d.Config.GameStateTime[d.gameState]),
		WildCard:       d.WildCard.Card,
		ShowCard:       d.ShowCard.Card,
		CardsNum:       int32(d.CardMgr.GetLeftCardCount()),
		OperSeatId:     d.OperateId,
		BankerSeatId:   d.BankerId,
		KingSeatId:     d.KingId,
		FirstSeatId:    d.FristOutId,
		UserSeatId:     p.seatPos,
	}
	for _, v := range d.seatPlayers {
		deskInfo.PlayersInfo = append(deskInfo.PlayersInfo, protocol.EnterDeskInfo{
			SeatPos:  v.seatPos,
			Nickname: v.name,
			Sex:      v.sex,
			HeadUrl:  v.head,
			StarNum:  v.starNum,
			IsBanker: v.isBanker,
			Sitdown:  v.sitdown,
			LiXian:   v.disconnect,
			IsKing:   v.IsKing,
			Show:     v.showed,
			Coins:    v.Coins,
		})
	}
	//游戏开始，自己不在游戏中
	deskInfo.PublicCard = d.GetPublicCard().Card
	return deskInfo
}
func (this *Desk) start(interface{}) {
	//检测桌子开始条件是否符合
	if len(this.players) < this.Config.MinPlayerNum {
		if len(this.players) == 0 {
			//解散桌子
		}
		this.gameState = GameStateWaitJoin
		this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
			GameState: int32(this.gameState),
		})
		return
	}
	log.Println("游戏开始！//////////////////////////////////")
	//游戏开始
	this.InitDesk()
	this.gameState = GameStateSendCard
	for k, v := range this.seatPlayers {
		this.doingPlayers[k] = v
		v.isJoin = true
	}
	//初始化玩家
	for _, v := range this.doingPlayers {
		v.InitPlayer()
	}
	//洗牌
	this.CardMgr.Shuffle()
	//发牌
	log.Println("发牌!/////////////////////////////////////")
	for i, p := range this.doingPlayers {
		if i == 0 && len(this.Config.FirstPlayerCards) == 13 {
			p.HandCards = append([]GCard{}, this.Int32ToGCard(this.Config.FirstPlayerCards)...)
		} else {
			p.HandCards = append([]GCard{}, this.CardMgr.SendCard(13)...)
		}
		this.CardMgr.QuickSortCLV(p.HandCards)
		//拆成牌组
		t := 0
		tcard := p.HandCards[0]
		for k, v := range p.HandCards {
			if v.GetCardColor() != tcard.GetCardColor() {
				p.CardsSet = append(p.CardsSet, protocol.CardsSet{Cards: this.GCardToInt32(p.HandCards[t:k])})
				t = k
				tcard = v
			}
		}
		log.Println(p.name, p.GetHandCardsString())
	}
	//定庄
	if this.BankerId == -1 {
		for i := 0; i < this.Config.FullPlayerNum; i++ {
			if this.doingPlayers[int32(i)] != nil {
				this.BankerId = this.doingPlayers[int32(i)].seatPos
			}
		}
	} else if this.isFirstRound {
		this.isFirstRound = false
	} else {
		for i := 0; i < this.Config.FullPlayerNum; i++ {
			bankernext := (this.BankerId + int32(i)) % int32(this.Config.FullPlayerNum)
			if this.doingPlayers[bankernext] != nil {
				this.BankerId = bankernext
			}
		}
	}
	//定首家(庄家的下家)
	for i := 1; i < this.Config.FullPlayerNum; i++ {
		bankernext := (this.BankerId + int32(i)) % int32(this.Config.FullPlayerNum)
		if this.doingPlayers[bankernext] != nil {
			this.FristOutId = bankernext
			this.OperateId = bankernext
			break
		}
	}
	log.Print("庄家：", this.doingPlayers[this.BankerId].name)
	log.Print("首家：", this.doingPlayers[this.FristOutId].name)
	//广播开始通知
	for k := range this.players {
		msg := this.GetDeskInfo(this.players[k])
		this.players[k].session.Push(PushGameStrat, &msg)
	}

	//定万能牌
	if this.Config.WildCard != 0 {
		this.WildCard = GCard{Poker.CardBase{Card: this.Config.WildCard}}
	} else {
		this.WildCard = this.CardMgr.SendCard(1)[0]
	}

	//翻第一张牌
	this.PublicCard = this.CardMgr.SendCard(1)
	log.Print("万能牌：", this.WildCard.Name)
	log.Println("第一张公摊牌：", this.PublicCard[0].Name)
	//发牌动画
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
		TotalTime: int32(this.Config.GameStateTime[this.gameState]),
	})
	for _, p := range this.doingPlayers {
		var handcard []int32
		for _, v := range p.HandCards {
			handcard = append(handcard, v.Card)
		}
		p.session.Push(PushSendCard, &protocol.GSendCardNotice{
			HandCards: handcard,
			WildCard:  this.WildCard.Card,
			FristCard: this.GetPublicCard().Card,
		})
	}
	this.ClearTimer()
	this.AddTimer(GameStateSendCard, this.Config.GameStateTime[GameStateSendCard], this.OpertionNotice, nil)

}

//玩家操作通知
func (this *Desk) OpertionNotice(interface{}) {
	if this.CardMgr.GetLeftCardCount() <= 0 {
		//广播游戏结束
		this.gameState = GameStateAbort
		this.ClearTimer()
		this.AddTimer(GameStateAbort, this.Config.GameStateTime[GameStateAbort]-4, this.GameEnd, nil)
		this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
			GameState: int32(this.gameState),
			Time:      int32(this.GetTimerNum(this.gameState)),
			TotalTime: int32(this.Config.GameStateTime[this.gameState]),
		})
		this.group.Broadcast(NoticeGameWin, &protocol.GPlayerWinNotice{
			SeatId:      -1,
			WinCoins:    0,
			SettleCoins: this.SettleCoins,
		})
		return
	}
	if this.OperateId == this.FristOutId {
		this.round++
	}
	this.gameState = GameStatePlay
	this.ClearTimer()
	this.AddTimer(GameStatePlay, this.Config.GameStateTime[GameStatePlay], this.PlayeroperOutTime, nil)
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
		TotalTime: int32(this.Config.GameStateTime[this.gameState]),
		OperateId: this.OperateId,
	})
	p := this.doingPlayers[this.OperateId]
	_, coins := this.GetGiveUpCoins(p)

	if this.round <= 1 {
		this.group.Broadcast(PushRoundInfo, &protocol.GRoundInfoNotice{GiveUpCoins: coins})
	} else {
		p.session.Push(PushRoundInfo, &protocol.GRoundInfoNotice{GiveUpCoins: coins})
	}
	log.Println("当前轮到：", this.doingPlayers[this.OperateId].name)
}

//玩家操作超时
func (this *Desk) PlayeroperOutTime(interface{}) {
	p := this.doingPlayers[this.OperateId]
	log.Printf("玩家%v,超时次数:%v\n", p.name, p.Timeout)
	if p == nil {
		return
	}
	handcard := p.HandCards
	if len(handcard) < 13 || len(handcard) > 14 {
		log.Error("座位号:%d  手牌数不对，手牌是%v", this.OperateId, handcard)
		return
	}
	if p.Timeout >= 1 || this.ShowPlayer != -1 { //玩家超时弃牌
		if this.ShowPlayer != -1 {
			this.PublicCard = append(this.PublicCard, this.ShowCard)
			this.ShowCard = GCard{}
			this.group.Broadcast(NoticeGameOperCard, &protocol.GPlayerOperNotice{
				SeatId:     p.seatPos,
				Opertion:   5,
				OperCard:   this.GetPublicCard().Card,
				PublicCard: this.GetPublicCard().Card,
				ShowCard:   0,
				CardsNum:   int32(this.CardMgr.GetLeftCardCount()),
			})
		}
		log.Println("玩家操作超时弃牌！")
		_ = this.GiveUp(p, true, &protocol.GGiveUpRequect{
			Cards: [][]int32{this.GCardToInt32(p.HandCards)},
		})
		this.ShowCard = GCard{}
		this.ShowPlayer = -1
		return
	}
	if len(handcard) == 13 {
		//摸牌操作
		this.OperCard(p, &protocol.GOperCardRequest{Opertion: 2}, true)
		this.ClearTimer()
		this.AddTimer(GameStatePlay, GameStatePlayAVTime, this.PlayeroperOutTime, nil)
		return
	}
	//出牌操作
	if len(handcard) == 14 {
		log.Println("牌组：", p.CardsSet)
		var tcards [][]GCard
		card := p.HandCards[len(p.HandCards)-1].Card
		x := -1
		y := -1
		for i, v := range p.CardsSet {
			var s []GCard
			for j, v1 := range v.Cards {
				if v1 == card && (x == -1 || len(p.CardsSet[x].Cards) > len(v.Cards)) {
					x = i
					y = j
				}
				s = append(s, GCard{Poker.CardBase{Card: v1}})
			}
			if len(s) > 0 {
				tcards = append(tcards, s)
			}
		}
		if x != -1 && y != -1 {
			tcards[x] = append(tcards[x][:y], tcards[x][y+1:]...)
			if len(tcards[x]) == 0 {
				tcards = append(tcards[:x], tcards[x+1:]...)
			}
		}
		IsHu, _ := this.CardMgr.CheckoutHu(tcards, this.WildCard)
		if IsHu {
			this.OperCard(p, &protocol.GOperCardRequest{
				Opertion: 4,
				OperCard: p.HandCards[len(p.HandCards)-1].Card,
			}, true)
		} else {
			this.OperCard(p, &protocol.GOperCardRequest{
				Opertion: 3,
				OperCard: p.HandCards[len(p.HandCards)-1].Card,
			}, true)
		}

		p.Timeout++
		p.deposit = true
	}
	log.Printf("玩家%v,超时次数:%v\n", p.name, p.Timeout)
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
		log.Debug("玩家%s手牌数%s不对!  ", p.name, len(p.HandCards), p.uid)
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
		//公摊牌是万能牌不能摸哦
		operCard := this.GetPublicCard()
		wildCard := this.WildCard
		if wildCard.GetCardColor() == Poker.CARD_COLOR_King {
			wildCard.Card = Poker.Card_Fang_1
		}
		if (operCard.GetCardValue() == wildCard.GetCardValue() || operCard.GetCardColor() == Poker.CARD_COLOR_King) &&
			len(this.PublicCard) > 1 {
			return p.session.Response(&protocol.GOperCardResponse{
				Opertion: 0,
				Error:    "公摊牌是万能牌不能摸哦！",
			})
		}
		//公摊牌不是万能牌
		p.HandCards = append(p.HandCards, operCard)
		this.PublicCard = append([]GCard{}, this.PublicCard[:len(this.PublicCard)-1]...)
		msgResponse.OperCard = operCard.Card
		msgResponse.PublicCard = this.GetPublicCard().Card
		msgResponse.Error = ""
		msgNotice.OperCard = operCard.Card
		msgNotice.PublicCard = this.GetPublicCard().Card
	}
	//摸牌堆的牌
	if msg.Opertion == 2 && len(p.HandCards) == 13 {
		operCard := this.CardMgr.SendCard(1)
		p.HandCards = append(p.HandCards, operCard[0])
		msgResponse.OperCard = operCard[0].Card
		msgResponse.PublicCard = this.GetPublicCard().Card
		msgResponse.Error = ""
		msgNotice.OperCard = 0
		msgNotice.PublicCard = this.GetPublicCard().Card
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
		msgResponse.PublicCard = this.GetPublicCard().Card
		msgResponse.Error = ""
		msgNotice.OperCard = msg.OperCard
		msgNotice.PublicCard = this.GetPublicCard().Card
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
		msgResponse.PublicCard = this.GetPublicCard().Card
		msgResponse.Error = ""
		msgNotice.OperCard = msg.OperCard
		msgNotice.PublicCard = this.GetPublicCard().Card
		this.ShowCard = GCard{Poker.CardBase{Card: msg.OperCard}}
		this.ShowPlayer = p.seatPos
	}
	msgResponse.CardsNum = int32(this.CardMgr.GetLeftCardCount())
	msgNotice.CardsNum = int32(this.CardMgr.GetLeftCardCount())
	if msgResponse.Error != "" {
		msgResponse.Opertion = 0
	} else {
		log.Printf("玩家：%v,操作：%v,牌: %v\n", p.name, msg.Opertion, msgResponse.OperCard)
	}
	if IsOuttime {
		err = p.session.Push(PushOperOutTime, &msgResponse)
	} else {
		p.Timeout = 0
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
		this.ClearTimer()
		this.gameState = GameStatePlayAV
		this.AddTimer(GameStatePlayAV, GameStatePlayAVTime, this.OutCard, nil)
	}
	return
}

//出牌动画
func (this *Desk) OutCard(interface{}) {
	//轮到下一位玩家
	for i := 1; i < this.Config.FullPlayerNum; i++ {
		nextplayer := (this.OperateId + int32(i)) % int32(this.Config.FullPlayerNum)
		if this.doingPlayers[nextplayer] != nil {
			this.OperateId = nextplayer
			break
		}
	}
	this.OpertionNotice(nil)
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
	log.Printf("整理牌请求：%v ", msg)
	tset, msgResponse.TotalPoint = this.HandleCardsSet(msg.CardsSets)
	p.CardsSet = tset
	msgResponse.CardsSets = tset
	msgResponse.Success = true
	msgResponse.Phase = msg.Phase
	if msgResponse.TotalPoint == 0 {
		msgResponse.IsHu = true
	}
	log.Printf("整理牌应答：%v ", msgResponse)
	err := p.session.Response(&msgResponse)
	//判断是否胡牌
	if this.ShowPlayer != p.seatPos || msg.Phase == Unspecified || (msg.Phase == Show && msgResponse.TotalPoint != 0) {
		return err
	}
	if msgResponse.TotalPoint != 0 {
		log.Println("玩家show牌失败弃牌！")
		this.PublicCard = append(this.PublicCard, this.ShowCard)
		this.group.Broadcast(NoticeGameOperCard, &protocol.GPlayerOperNotice{
			SeatId:     p.seatPos,
			Opertion:   5,
			OperCard:   this.GetPublicCard().Card,
			PublicCard: this.GetPublicCard().Card,
			ShowCard:   0,
			CardsNum:   int32(this.CardMgr.GetLeftCardCount()),
		})
		this.ClearTimer()
		_ = this.GiveUp(p, true, &protocol.GGiveUpRequect{Cards: this.CardsSetToInt32(p.CardsSet)})
		this.ShowCard = GCard{}
		this.ShowPlayer = -1
		return err
	}
	//添加结算时间
	this.ClearTimer()
	this.AddTimer(GameStateStettle, this.Config.GameStateTime[GameStateStettle]-1, this.SettleTimeOut, nil)
	this.gameState = GameStateStettle
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
		TotalTime: int32(this.Config.GameStateTime[this.gameState]),
	})
	this.WinPlayerId = p.seatPos
	//广播
	p.session.Push(PushWinGame, &protocol.GGiveUpResponse{
		Success:     true,
		Coins:       0,
		PlayerCoins: p.Coins,
		PointNum:    p.Point,
		TotalCoins:  this.SettleCoins,
	})
	log.Printf(p.name, "赢了！！！\n")
	this.group.Multicast(NoticeGameWin, &protocol.GPlayerWinNotice{
		SeatId:      this.WinPlayerId,
		SettleCoins: this.SettleCoins,
	}, func(s *session.Session) bool {
		if s == p.session {
			return false
		}
		return true
	})
	return err
}

//结算超时
func (this *Desk) SettleTimeOut(interface{}) {
	//检测玩家是否都结算了
	for _, v := range this.doingPlayers {
		if v.seatPos == this.WinPlayerId || v.settle {
			continue
		}
		_ = this.Settle(v, &protocol.GSettleRequect{
			Cards: this.CardsSetToInt32(v.CardsSet),
		}, true)
	}
}

//结算
func (this *Desk) Settle(p *Player, msg *protocol.GSettleRequect, isOutTime bool) (err error) {
	//检测
	if p == nil {
		log.Debug("Desk.Settle: *Player is nil!")
		return
	}
	if this.gameState != GameStateStettle {
		err = p.session.Response(&protocol.GSettleResponse{
			LoseCoins:   0,
			Error:       "状态错误！",
			SettleCoins: this.SettleCoins,
			MyCoins:     p.Coins,
		})
		return
	}
	//如果是赢的玩家就不用结算
	if this.WinPlayerId == p.seatPos {
		err = p.session.Response(&protocol.GSettleResponse{
			LoseCoins:   0,
			Error:       "消息错误！",
			SettleCoins: this.SettleCoins,
			MyCoins:     p.Coins,
		})
		return
	}
	if p.settle {
		err = p.session.Response(&protocol.GSettleResponse{
			LoseCoins:   0,
			Error:       "不能重复结算！",
			SettleCoins: this.SettleCoins,
			MyCoins:     p.Coins,
		})
		return
	}

	//取出牌组
	var tcards [][]GCard
	var show []GCard
	for _, v := range msg.Cards {
		var s []GCard
		for _, v1 := range v {
			s = append(s, GCard{Poker.CardBase{Card: v1}})
		}
		show = append(show, s...)
		tcards = append(tcards, s)
	}
	//检验发过来的牌组是否都是玩家的手牌
	this.CardMgr.QuickSortCLV(show)
	this.CardMgr.QuickSortCLV(p.HandCards)
	for k, v := range p.HandCards {
		if show[k].Card != v.Card {
			err = p.session.Response(&protocol.GSettleResponse{
				LoseCoins:   0,
				Error:       "发过来的牌组错误！",
				SettleCoins: this.SettleCoins,
				MyCoins:     p.Coins,
			})
			return
		}
	}
	//
	nextplayerId := int32(0)
	for i := 1; i < this.Config.FullPlayerNum; i++ {
		nextplayerId = (this.OperateId + int32(i)) % int32(this.Config.FullPlayerNum)
		if this.doingPlayers[nextplayerId] != nil {
			break
		}
	}
	if this.round <= 1 && (this.WinPlayerId == this.FristOutId || this.WinPlayerId == nextplayerId) { //天胡或者地胡
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
	p.isJoin = false
	//添加记录
	this.GameRecord.EndInfo = append(this.GameRecord.EndInfo, protocol.PlayerEndInfo{
		Name:      p.name,
		Head:      p.head,
		CardsSets: p.CardsSet,
		Point:     p.Point,
		Coins:     p.win,
	})
	//结算通知
	if isOutTime {
		p.session.Push(PushSettle, &protocol.GSettleResponse{
			PointNum:    p.Point,
			LoseCoins:   -1 * p.win,
			SettleCoins: this.SettleCoins,
			MyCoins:     p.Coins,
		})
	} else {
		err = p.session.Response(&protocol.GSettleResponse{
			PointNum:    p.Point,
			LoseCoins:   -1 * p.win,
			SettleCoins: this.SettleCoins,
			MyCoins:     p.Coins,
		})
	}
	this.group.Multicast(NoticeSettle, &protocol.GSettleNotice{
		SeatId:      p.seatPos,
		LosingCoins: -1 * p.win,
		SettleCoins: this.SettleCoins,
		Point:       p.Point,
	}, func(s *session.Session) bool {
		if s == p.session {
			return false
		}
		return true
	})
	//判断是否都结算完了
	for _, v := range this.doingPlayers {
		if v.seatPos == this.WinPlayerId || v.settle {
			continue
		}
		return
	}
	winPlayer := this.doingPlayers[this.WinPlayerId]
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
	log.Println("玩家的结算信息：", this.GameRecord.EndInfo)
	this.gameState = GameStateStettleAV
	this.ClearTimer()
	this.AddTimer(GameStateStettleAV, GameStateStettleAVTime, this.SettleAV, nil)

	return
}
func (this *Desk) SettleAV(interface{}) {
	winPlayer := this.doingPlayers[this.WinPlayerId]
	this.gameState = GameStateEnd
	this.ClearTimer()
	this.AddTimer(GameStateEnd, this.Config.GameStateTime[GameStateEnd], this.GameEnd, nil)
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
		TotalTime: int32(this.Config.GameStateTime[this.gameState]),
	})
	winPlayer.session.Push(PushWinGame, &protocol.GGiveUpResponse{
		Success:     true,
		Coins:       this.SettleCoins,
		PlayerCoins: winPlayer.Coins,
		PointNum:    winPlayer.Point,
		TotalCoins:  0,
	})
	this.group.Broadcast(NoticeGameWin, &protocol.GPlayerWinNotice{
		SeatId:      winPlayer.seatPos,
		WinCoins:    this.SettleCoins,
		SettleCoins: 0,
	})
	this.group.Broadcast(NoticeEndInfo, &this.GameRecord)
}

//放弃
func (this *Desk) GiveUp(p *Player, isOutTime bool, msg *protocol.GGiveUpRequect) error {
	//不是自己的回合不能弃牌
	if this.OperateId != p.seatPos {
		return p.session.Response(&protocol.GGiveUpResponse{
			Success: false,
		})
	}
	//检测牌是否是自己的手牌
	var tCard []int32
	for _, v := range msg.Cards {
		tCard = append(tCard, v...)
	}
	Gcard := this.Int32ToGCard(tCard)
	this.CardMgr.QuickSortCLV(Gcard)
	this.CardMgr.QuickSortCLV(p.HandCards)
	log.Println(p.HandCards, Gcard)
	for k, v := range p.HandCards {
		if Gcard[k].Card != v.Card {
			return p.session.Response(&protocol.GGiveUpResponse{
				Success: false,
			})
		}
	}
	p.CardsSet = this.Int32ToCardsSet(msg.Cards)
	//扣钱
	this.ClearTimer()
	var coins int64
	p.Point, coins = this.GetGiveUpCoins(p)
	this.SettleCoins += coins
	p.win = -1 * coins
	p.Coins -= coins
	p.settle = true
	p.CardsSet = []protocol.CardsSet{{
		Cards: make([]int32, 13),
		Type:  0,
		Point: p.Point,
	}}
	//添加记录
	this.GameRecord.EndInfo = append(this.GameRecord.EndInfo, protocol.PlayerEndInfo{
		Name:      p.name,
		Head:      p.head,
		CardsSets: p.CardsSet,
		Point:     p.Point,
		Coins:     p.win,
	})
	//
	msgNotice := protocol.GGiveUpNotice{}
	msgNotice.IsShow = isOutTime
	msgNotice.LosingCoins = coins
	msgNotice.SeatId = p.seatPos
	msgNotice.SettleCoins = this.SettleCoins
	msgNotice.Point = p.Point
	var err error
	if isOutTime {
		p.session.Push(PushLoseGame, &protocol.GGiveUpResponse{
			Success:     true,
			Coins:       coins,
			PlayerCoins: p.Coins,
			PointNum:    p.Point,
			TotalCoins:  this.SettleCoins,
		})
	} else {
		err = p.session.Response(&protocol.GGiveUpResponse{
			Success:     true,
			Coins:       coins,
			PlayerCoins: p.Coins,
			PointNum:    p.Point,
			TotalCoins:  this.SettleCoins,
		})
	}
	//广播玩家弃牌
	log.Println("广播玩家弃牌!!!!!!!!!!")
	_ = this.group.Multicast(NoticeGiveUp, &msgNotice, func(s *session.Session) bool {
		if s == p.session && isOutTime {
			return false
		}
		return true
	})
	p.isJoin = false
	delete(this.doingPlayers, p.seatPos)
	//如果是轮到自己操作的时候放弃，那么切换下一个玩家
	if this.OperateId == p.seatPos {
		//轮到下一位玩家
		for i := 1; i < this.Config.FullPlayerNum; i++ {
			nextplayer := (this.OperateId + int32(i)) % int32(this.Config.FullPlayerNum)
			if this.doingPlayers[nextplayer] != nil {
				this.OperateId = nextplayer
				break
			}
		}
	}
	//如果只剩一家就结算
	if len(this.doingPlayers) == 1 {
		//广播
		log.Println("玩家不战而胜！！！！！")
		this.WinPlayerId = this.OperateId
		winPlayer := this.doingPlayers[this.WinPlayerId]
		winPlayer.Coins += this.SettleCoins
		winPlayer.win = this.SettleCoins
		winPlayer.Point = 0
		this.GameRecord.EndInfo = append([]protocol.PlayerEndInfo{{
			Name:      winPlayer.name,
			Head:      winPlayer.head,
			CardsSets: winPlayer.CardsSet,
			Point:     0,
			Coins:     winPlayer.win,
		}}, this.GameRecord.EndInfo...)
		this.gameState = GameStateAbort
		this.ClearTimer()
		this.AddTimer(GameStateAbort, this.Config.GameStateTime[GameStateAbort], this.GameEnd, nil)
		this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
			GameState: int32(this.gameState),
			Time:      int32(this.GetTimerNum(this.gameState)),
			TotalTime: int32(this.Config.GameStateTime[this.gameState]),
		})
		winPlayer.session.Push(PushWinGame, &protocol.GGiveUpResponse{
			Success:     true,
			Coins:       this.SettleCoins,
			PlayerCoins: winPlayer.Coins,
			PointNum:    winPlayer.Point,
			TotalCoins:  0,
		})
		this.group.Multicast(NoticeGameWin, &protocol.GPlayerWinNotice{
			SeatId:      this.WinPlayerId,
			WinCoins:    this.SettleCoins,
			SettleCoins: 0,
		}, func(s *session.Session) bool {
			if s == winPlayer.session {
				return false
			}
			return true
		})
		this.group.Broadcast(NoticeEndInfo, &this.GameRecord)
		return err
	}
	this.ClearTimer()
	this.gameState = GameStatePlayAV
	this.AddTimer(GameStatePlayAV, GameStatePlayAVTime, this.OpertionNotice, nil)
	return err
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

	log.Println("玩家下线了：uid", s.UID())

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
	//不在座位上设为-1
	p.sitdown = false
	//设置座位号
	for i := int32(0); i < int32(this.Config.FullPlayerNum); i++ {
		if this.seatPlayers[i] == nil {
			this.seatPlayers[i] = p
			p.SetSeatPos(i)
			p.sitdown = true
			break
		}
	}

	for _, p := range this.players {
		p.setDesk(this)
	}
	playernum := len(this.seatPlayers)
	//如果玩家玩家人数够了
	if playernum == 1 {
		p.IsKing = true
		this.KingId = p.seatPos
		this.BankerId = p.seatPos
	}
	return true
}

//游戏结束
func (this *Desk) GameEnd(interface{}) {
	log.Println("游戏结束！/////////////////////////////")
	//清理桌子内的玩家
	for _, v := range this.players {
		v.isJoin = false
		if v.disconnect || !v.sitdown || v.Coins < 8000 {
			this.ExitDesk(v)
			v.ExitDesk()
		}
	}
	this.ClearTimer()
	//判断是否人数足够重开一把
	if len(this.seatPlayers) < this.Config.MinPlayerNum {
		this.gameState = GameStateWaitJoin
	} else {
		//游戏重开一局
		this.gameState = GameStateWaitStart
		this.AddTimer(GameStateWaitStart, this.Config.GameStateTime[GameStateWaitStart], this.start, nil)
	}
	this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
		GameState: int32(this.gameState),
		Time:      int32(this.GetTimerNum(this.gameState)),
		TotalTime: int32(this.Config.GameStateTime[this.gameState]),
	})
}

func (this *Desk) Int32ToGCard(str []int32) []GCard {
	var tset []GCard
	for _, v := range str {
		card := GCard{Poker.CardBase{
			Card: v,
		}}
		card.GetCardName()
		tset = append(tset, card)
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
func (this *Desk) Int32ToCardsSet(cardsSet [][]int32) []protocol.CardsSet {
	var str []protocol.CardsSet
	for _, v := range cardsSet {
		str = append(str, protocol.CardsSet{Cards: v})
	}
	return str
}

//整理牌组的类型
func (this *Desk) HandleCardsSet(CardsSets []protocol.CardsSet) ([]protocol.CardsSet, int32) {
	//计算这些牌的组合情况
	totalPoint := int32(0)
	sets := append([]protocol.CardsSet{}, CardsSets...)
	have1st := -1
	have2st := -1
	var firstK []int
	var secondK []int
	var GroupK []int
	var Other []int32
	//找出第一生命和第二生命
	for k, v := range CardsSets {
		if this.CardMgr.Is1stLife(this.Int32ToGCard(v.Cards)) {
			sets[k].Type = Type1stLife
			sets[k].Point = 0
			firstK = append(firstK, k)
		} else if this.CardMgr.Is2stLife(this.Int32ToGCard(v.Cards), this.WildCard) {
			sets[k].Type = Type2stLife
			sets[k].Point = 0
			secondK = append(secondK, k)
		} else if this.CardMgr.IsSetLife(this.Int32ToGCard(v.Cards), this.WildCard) {
			sets[k].Type = TypeSet
			sets[k].Point = 0
			GroupK = append(GroupK, k)
		} else {
			sets[k].Type = TypeOther
			Other = append(Other, v.Cards...)
			sets[k].Point = int32(this.CardMgr.ComputePoint([][]GCard{this.Int32ToGCard(v.Cards)}, this.WildCard, have1st != -1))
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
		if v == have1st {
			continue
		}
		sets[v].Type = Type2stLife
		secondK = append(secondK, v)
	}
	//确定唯一第二生命
	for _, v := range secondK {
		if have1st == -1 {
			sets[v].Type = TypeNeed1stLife
			have2st = -1
			sets[v].Point = int32(this.CardMgr.ComputePoint([][]GCard{this.Int32ToGCard(sets[v].Cards)}, this.WildCard, have1st != -1))
			Other = append(Other, sets[v].Cards...)
			continue
		}
		if have2st == -1 {
			have2st = v
			continue
		}
		if have2st == v {
			continue
		}
		sets[v].Type = TypeSequence
	}
	//确定集合
	for _, v := range GroupK {
		if have2st == -1 {
			sets[v].Type = TypeNeed2stLife
			sets[v].Point = int32(this.CardMgr.ComputePoint([][]GCard{this.Int32ToGCard(sets[v].Cards)}, this.WildCard, have1st != -1))
			Other = append(Other, sets[v].Cards...)
		}
	}
	//计算分数
	totalPoint = int32(this.CardMgr.ComputePoint([][]GCard{this.Int32ToGCard(Other)}, this.WildCard, have1st != -1))
	if totalPoint > this.Config.MaxLosePoint {
		totalPoint = this.Config.MaxLosePoint
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
	tGCard := append([]GCard{this.WildCard}, this.PublicCard...)
	this.CardMgr.QuickSortCLV(tGCard)
	//花色区分
	for _, v := range tGCard {
		if v.GetCardColor() == Poker.CARD_COLOR_Fang {
			cards[0] = append(cards[0], v.GetCardValue())
			continue
		}
		if v.GetCardColor() == Poker.CARD_COLOR_Mei {
			cards[1] = append(cards[1], v.GetCardValue())
			continue
		}
		if v.GetCardColor() == Poker.CARD_COLOR_Hong {
			cards[2] = append(cards[2], v.GetCardValue())
			continue
		}
		if v.GetCardColor() == Poker.CARD_COLOR_Hei {
			cards[3] = append(cards[3], v.GetCardValue())
			continue
		}
	}

	return p.session.Response(&protocol.GOutCardRecordResponse{
		Success:    true,
		CardRecord: cards,
	})

}

//获取玩家弃牌输的点值和金额
func (this *Desk) GetGiveUpCoins(p *Player) (point int32, coins int64) {
	//扣钱
	if this.ShowPlayer == p.seatPos {
		point = this.Config.MaxLosePoint
	} else if this.round <= 1 {
		point = this.Config.Round1GiveUpPoint
	} else if this.round == 2 {
		point = this.Config.Round2GiveUpPoint
	} else {
		point = this.Config.MaxLosePoint
	}
	coins = this.PointValue * int64(point)
	return
}

//获取公摊牌
func (this *Desk) GetPublicCard() GCard {
	if len(this.PublicCard) > 0 {
		return this.PublicCard[len(this.PublicCard)-1]
	}
	return GCard{CardBase: Poker.CardBase{Card: 0}}
}

//玩家退出
func (this *Desk) ExitDesk(p *Player) bool {
	IsLevel := false
	playerState := int32(0)
	if this.gameState < GameStateSendCard || this.gameState == GameStateEnd || this.gameState == GameStateAbort || !p.isJoin || p.settle {
		IsLevel = true
	}
	if IsLevel {
		delete(this.doingPlayers, p.seatPos)
		delete(this.seatPlayers, p.seatPos)
		for k := range this.players {
			if this.players[k] == p {
				this.players = append(this.players[:k], this.players[k+1:]...)
				break
			}
		}
		playerState = PlayerStateLeave
	} else {
		p.disconnect = true
		playerState = PlayerStateLixian
	}
	this.group.Leave(p.session)
	this.group.Broadcast(NoticePlayerState, &protocol.PlayerStateNotice{
		SeatId:      p.seatPos,
		PlayerState: playerState,
	})
	if playerState == PlayerStateLeave &&
		this.gameState == GameStateWaitStart &&
		len(this.seatPlayers) < this.Config.MinPlayerNum {
		//广播游戏状态
		this.ClearTimer()
		this.gameState = GameStateWaitJoin
		this.group.Broadcast(NoticeGameState, &protocol.GGameStateNotice{
			GameState: int32(this.gameState),
			Time:      int32(this.GetTimerNum(this.gameState)),
			TotalTime: int32(this.Config.GameStateTime[this.gameState]),
			OperateId: p.seatPos,
		})
	}

	return true
}
