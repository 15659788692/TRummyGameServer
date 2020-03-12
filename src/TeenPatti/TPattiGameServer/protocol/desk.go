package protocol


type EnterDeskInfo struct {


	SeatPos   int    `json:"seatPos"`       //玩家的坐位

	Nickname  string `json:"nickname"`      //玩家的名称
	Sex       int    `json:"sex"`           //性别
	HeadUrl   string `json:"headURL"`       //图标

	Score     int    `json:"score"`         //玩家的余额
	StarNum   int    `json:"starNum"`       //星级个数


	IsBanker  bool   `json:"isBanker"`      //是否是庄家
	Sitdown   bool    `json:"sitdown"`      //是否坐下
	Betting   bool    `json:"betting"`      //当前是否投注中


	Packed    bool    `json:"Packed"`       //界面上三个按钮的状态
	Show      bool    `json:"Show"`         //是示请求SHOW的按钮状态
	Blind     bool    `json:"Blind"`        //Blink按钮状态

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

type PlayerEnterDesk struct {
	Data []EnterDeskInfo `json:"data"`
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

	deskId   string   `json:"deskId"`        //桌子的ID

	Status   int32    `json:"deskStatus"`   //桌子的基本状态,值为constant包中的状态值

	CurTime  int64    `json:"curTime"`      //系统的当前时间

	KeepSecond     int      `json："keepSecond"`    //此状态保持的时间，秒为单位0:为不限

}



type  PlayScore struct {
	NickName   string  `json:"nickName"`
	Score      int     `json:"score"`

}


//服务器会通知玩家余额
type PlaysScoreInfo struct {

     scoreInfos  [] PlayScore
}

//通知谁当庄
type  BroadBankerInfo struct {

	NickName string  `json:"nickName"`   //当前当庄的人
	SeatPos  int     `json:"seatPos"`    //庄家的坐位
}


//通知哪个玩家可以投注了
type  BroadPlayerBetAcitve struct {

	NickName     string  `json:"nickName"`   //当前可投注的人
	SeatPos      int     `json:"seatPos"`    //此人的坐位

    KeepTime     int     `json:"keepTime"`   //可投注的秒数
    LostTime     int     `json:"lostTime"`   //已经过的秒数
}

//通知有玩家下注了,客户端需要飞筹码动画
type  BroadPlayerBetting struct {

	NickName     string  `json:"nickName"`   //玩家名称

	SeatPos      int     `json:"seatPos"`    //此人的坐位
	BetCount     int32   `json:"betCount"`   //下注的额度
}


//服务器通知玩家坐下或站起来
type BroadsPlayerSitdown struct {

	NickName     string  `json:"nickName"`   //玩家名称

	SeatPos      int     `json:"seatPos"`    //此人的坐位
	Sitdown      bool    `json:"sitdown"`    //坐下状态true,坐下，false 站起来
}

//通知指定玩家起立状态
type NotifyPlayerSitdown struct {

	Uid          int64

	NickName     string  `json:"nickName"`   //玩家名称
	SeatPos      int     `json:"seatPos"`    //此人的坐位
	Sitdown      bool    `json:"sitdown"`    //坐下状态true,坐下，false 站起来
}
