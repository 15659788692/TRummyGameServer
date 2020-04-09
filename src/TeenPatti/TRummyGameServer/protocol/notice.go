package protocol

//某玩家进入桌子信息
type EnterDeskInfo struct {
	SeatPos  int32  `json:"seatPos"`  //玩家的坐位
	Nickname string `json:"nickname"` //玩家的名称
	Sex      int    `json:"sex"`      //性别
	HeadUrl  string `json:"headURL"`  //图标
	Score    int    `json:"score"`    //玩家的余额
	StarNum  int    `json:"starNum"`  //星级个数
	IsBanker bool   `json:"isBanker"` //是否是庄家
	IsKing   bool   `json:"isKing`    //是否是房主
	Sitdown  bool   `json:"sitdown"`  //是否坐下
	LiXian   bool   `json:"liXian"`   //是否离线
	Show     bool   `json:"show"`     //是否showed
	IsSettle bool   `json:"isSettle"` //是否结算完
	Coins    int64  `json:"coins"`    //玩家的金额
}

//游戏状态通知
type GGameStateNotice struct {
	GameState int32 //状态
	Time      int32 //倒计时
	TotalTime int32 //总时间
	OperateId int32 //操作状态下操作的玩家
}

//游戏开始通知
//type GGameStartNotice struct {
//	BankerId int32 //座位号
//	FristID  int32 //首出玩家的座位号
//
//}

//发牌
type GSendCardNotice struct {
	HandCards []int32
	WildCard  int32
	FristCard int32
}

//回合桌子信息
type GRoundInfoNotice struct {
	GiveUpCoins int64
}

//操作广播
type GPlayerOperNotice struct {
	SeatId     int32 //座位号
	Opertion   int32 //摸牌操作	1.摸公摊牌   2.摸牌堆里的牌   3.出牌到牌堆   4.出牌到Show
	OperCard   int32 //操作的牌	如果Opertion是2，则OperCard为 0，
	PublicCard int32 //公摊牌	公摊牌的变化
	ShowCard   int32
	CardsNum   int32 //牌堆剩余张数
}

//胡牌广播
type GPlayerWinNotice struct {
	SeatId      int32 //座位号
	WinCoins    int64 //玩家赢的钱
	SettleCoins int64 //结算金额

}

//放弃广播
type GGiveUpNotice struct {
	SeatId      int32 //	玩家座位号
	LosingCoins int64 //	玩家输的金额
	SettleCoins int64 //	结算区的总金额
	Point       int32 //	输的点数
	IsShow      bool  //	是否是show
}

//广播结算
type GSettleNotice struct {
	SeatId      int32 //	玩家座位号
	LosingCoins int64 //	玩家输的金额
	SettleCoins int64 //	结算区的总金额
	Point       int32 //	输的点数
}

//结算界面广播
type GEndForm struct {
	EndInfo []PlayerEndInfo
}

type PlayerEndInfo struct {
	Name      string
	Head      string
	CardsSets []CardsSet
	Point     int32
	Coins     int64
}

//广播玩家状态
type PlayerStateNotice struct {
	SeatId      int32
	PlayerState int32 //1.加入房间 2.离开(空) 3.重连 4.掉线(空) 5.坐下 6.起立(空) 部分玩家信息为空
	PlayerInfo  EnterDeskInfo
}
