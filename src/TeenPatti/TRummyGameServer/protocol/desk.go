package protocol

//玩家加入后桌面的详细信息
type DeskInfo struct {
	PointValue    int64           `json:"pointValue"`    //点值，底注的意思，输一点代表80分
	DecksNum      int32           `json:"decksNum"`      //几副牌
	MaxWining     int64           `json:"maxWining"`     //最多赢多少分
	Maxlosing     int64           `json:"maxlosing"`     //最多输多少分
	SettleCoins   int64           `json:"settleCoins"`   //结算区金额
	GameState     int32           `json:"gameState"`     //游戏状态
	GameStateTime int32           `json:"gameStateTime"` //游戏状态时间
	WildCard      int32           `json:"wildCard"`      //万能牌
	ShowCard      int32           `json:"showCard"`      //show区的牌
	PublicCard    int32           `json:"publicCard"`    //公摊牌	公摊牌的变化
	CardsNum      int32           `json:"cardsNum"`      //牌堆剩余张数
	OperSeatId    int32           `json:"operSeatId"`    //当前操作的玩家座位号
	BankerSeatId  int32           `json:"bankerSeatId"`  //庄家的座位号
	KingSeatId    int32           `json:"kingSeatId"`    //房主的座位号
	FirstSeatId   int32           `json:"firstSeatId"`   //首出玩家的座位号
	UserSeatId    int32           `json:"userSeatId"`    //玩家自己的座位号
	PlayersInfo   []EnterDeskInfo `json:"playersInfo"`   //所有玩家的详细信息
}

type ExitRequest struct {
	IsDestroy bool `json:"isDestroy"`
}

type ExitResponse struct {
	AccountId int64 `json:"acid"`
	IsExit    bool  `json:"isexit"`
	ExitType  int   `json:"exitType"`
	DeskPos   int   `json:"deskPos"`
}

type DeskBasicInfo struct {
	DeskID string `json:"deskId"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
	Mode   int    `json:"mode"`
}

type ScoreInfo struct {
	Uid   int64 `json:"acId"`
	Score int   `json:"score"`
}

type DeskOptions struct {
	Mode     int `json:"mode"`
	MaxRound int `json:"round"`
	MaxFan   int `json:"maxFan"`
}

//玩家回到进入桌子消息后，需要发送这个请求给桌子
type ClientInitCompletedNotify struct {
	IsReEnter bool `json:"isReenter"`
}

//桌子的基本状态
type DeskStatusInfo struct {
	deskId string `json:"deskId"` //桌子的ID

	Status int32 `json:"deskStatus"` //桌子的基本状态,值为constant包中的状态值

	CurTime int64 `json:"curTime"` //系统的当前时间

	KeepSecond int `json："keepSecond"` //此状态保持的时间，秒为单位0:为不限

}

type PlayScore struct {
	NickName string `json:"nickName"`
	Score    int    `json:"score"`
}

//服务器会通知玩家余额
type PlaysScoreInfo struct {
	scoreInfos []PlayScore
}

//通知谁当庄
type BroadBankerInfo struct {
	NickName string `json:"nickName"` //当前当庄的人
	SeatPos  int    `json:"seatPos"`  //庄家的坐位
}

//通知哪个玩家可以投注了
type BroadPlayerBetAcitve struct {
	NickName string `json:"nickName"` //当前可投注的人
	SeatPos  int    `json:"seatPos"`  //此人的坐位

	KeepTime int `json:"keepTime"` //可投注的秒数
	LostTime int `json:"lostTime"` //已经过的秒数
}

//通知有玩家下注了,客户端需要飞筹码动画
type BroadPlayerBetting struct {
	NickName string `json:"nickName"` //玩家名称

	SeatPos  int   `json:"seatPos"`  //此人的坐位
	BetCount int32 `json:"betCount"` //下注的额度
}

//服务器通知玩家坐下或站起来
type BroadsPlayerSitdown struct {
	NickName string `json:"nickName"` //玩家名称

	SeatPos int  `json:"seatPos"` //此人的坐位
	Sitdown bool `json:"sitdown"` //坐下状态true,坐下，false 站起来
}

//通知指定玩家起立状态
type NotifyPlayerSitdown struct {
	Uid int64

	NickName string `json:"nickName"` //玩家名称
	SeatPos  int    `json:"seatPos"`  //此人的坐位
	Sitdown  bool   `json:"sitdown"`  //坐下状态true,坐下，false 站起来
}

//游戏部分
//操作牌
type GOperCardRequest struct {
	Opertion int32 //摸牌操作   1.摸公摊牌   2.摸牌堆里的牌   3.出牌到公摊区   4.出牌到Show	  5.弃牌
	OperCard int32 //获得的牌		125为0	34为牌值
}

type GOperCardResponse struct {
	Opertion   int32  //摸牌操作   -1.网络错误  0.操作错误  1.摸公摊牌   2.摸牌堆里的牌   3.出牌到公摊区   4.出牌到Show
	OperCard   int32  //操作的牌
	PublicCard int32  //公摊牌
	CardsNum   int32  //牌堆剩余牌的数量
	Error      string //报错信息
}

type GSetHandCardRequest struct {
	CardsSets []CardsSet
	Phase     int32
}

type CardsSet struct {
	Cards []int32
	Type  int32
	Point int32
}

type GSetHandCardResponse struct {
	Success    bool //是否成功
	CardsSets  []CardsSet
	TotalPoint int32
	IsHu       bool //是否胡牌
	Phase      int32
	Error      string //报错信息
}

//结算
type GSettleRequect struct {
	Cards [][]int32
}

type GSettleResponse struct {
	LoseCoins   int64
	PointNum    int32
	SettleCoins int64
	MyCoins     int64
	Error       string //报错信息
}

//放弃
type GGiveUpRequect struct {
	Cards [][]int32
}

type GGiveUpResponse struct {
	Success     bool //是否成功
	Coins       int64
	TotalCoins  int64 //桌子上的钱
	PlayerCoins int64
	PointNum    int32
}

//请求出牌记录
type GOutCardRecordRequect struct {
}

type GOutCardRecordResponse struct {
	Success    bool      //是否成功
	CardRecord [][]int32 //出牌记录
}
