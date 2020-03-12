package protocol


type EnterDeskInfo struct {

	DeskPos  int    `json:"deskPos"`
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	IsReady  bool   `json:"isReady"`
	Sex      int    `json:"sex"`
	IsExit   bool   `json:"isExit"`
	HeadUrl  string `json:"headURL"`
	Score    int    `json:"score"`
	IP       string `json:"ip"`
	Offline  bool   `json:"offline"`

}


//请求退出
type ExitRequest struct {
	IsDestroy bool `json:"isDestroy"`
}


//退出的回复
type ExitResponse struct {
	AccountId int64 `json:"acid"`
	IsExit    bool  `json:"isexit"`
	ExitType  int   `json:"exitType"`
	DeskPos   int   `json:"deskPos"`
}


//玩家进入桌子后的回复信息
type PlayerEnterDesk struct {
	Data []EnterDeskInfo `json:"data"`
}

//桌子的基本信息
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

