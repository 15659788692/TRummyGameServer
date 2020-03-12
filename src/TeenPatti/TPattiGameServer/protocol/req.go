package protocol

type ReJoinDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReJoinDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}


type ReEnterDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReEnterDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

//---------------------------------------------------------------------------------------------------


type TableInfo struct {

	DeskNo    string              `json:"deskId"`        //桌子的名称

	CreateTime  int64             `json:"createdTime"`   //桌子创建的时间

	playerNum     int             `json:"playerum"`      //玩家的个数

	BootAmout     int             `json:"bootAmout"`       //低注,进入此桌最少的投注额
	MaxBlinds     int             `json:"maxBlinds"`       //最大盲注,最多可盖牌的圈数
	ChaalLimit    int             `json:"chaalLimit"`      //单注限额
	PotLimit      int             `json:"potLimit"`        //单局总投注额度

	Status    int32               `json:"status"`          //此游戏桌子的状态,值为const.go内

	Round     uint32              `json:"round"`           //第几局

}

//--------------------------------------------------------------------

type JoinDeskRequest struct {

    Uid       int64     `json:"uid"`
	NickName  string    `json:"nickName"`
}


type JoinDeskResponse struct {

	Success    bool       `json:"success"`      //入桌是否成功

	TableInfo  TableInfo  `json:"tableInfo"`    //加入桌子的相关信息

	Code      int         `json:"code"`         //错误码
	Error     string      `json:"error"`        //若不成功，错误原因
}

//-----------------------------------------------------------------------

type DestoryDeskRequest struct {

	DeskNo string `json:"deskId"`
}

type PlayerOfflineStatus struct {
	Uid     int64 `json:"uid"`
	Offline bool  `json:"offline"`
}

//-------------------------------------------------------------------------

//投注请求
type PlayerBetRequest struct {

	Uid       int64   `json:"uId"`              //帐号ID
	BetCount  int32   `json:"betCount"`         //投注的额度
}

//投注的回复
type PlayerBetResponse struct {

	Uid       int64   `json:"uId"`               //帐号ID

	BetCount  int32   `json:"betCount"`          //此次投注额度的回复
    TotalBet  int64   `json:"totalBet"`          //玩家总的下注

	Success   bool    `json:"success"`          //操作是否成功,1成功,0不成功
	Error     string  `json:"error"`            //错误原因
}
//---------------------------------------------------------------------------

//按下看牌按钮
type  PlayerSeeRequest struct {

	Uid   int64   `json:"uId"`              //帐号ID
}

type  PlayerSeeResponse struct {

	Success   bool    `json:"success"`          //操作是否成功,1成功,0不成功

	Uid        int64   `json:"uId"`       //帐号ID
	pair [3]   byte    `json:"pair"`      //3张扑克牌
	Error     string  `json:"error"`            //错误原因
}

//-------------------------------------------------------------------------
//按下了Pack按钮
type  PlayerPackRequest struct {

	Uid       int64   `json:"uId"`   //帐号ID
}


//按钮的回复
type PlayerPackResonse struct {

	Uid       int64   `json:"uId"`      //帐号ID
	Enable    bool    `json:"enable"`   //操作后可用

	Success   bool    `json:"success"`  //操作是否成功,1成功,0不成功
	Error     string   `json:"error"`   //错误原因
}

//-------------------------------------------------------------------------
//按下了Show按钮
type PlayerShowRequest  struct {

	Uid    int64   `json:"uId"`   //帐号ID
}

//按下Show按钮后的回复
type PlayerShowResonse  struct {

	Uid    int64   `json:"uId"`  		 //帐号ID

	Enable    bool    `json:"enable"`   //操作后可用

	Success   bool    `json:"success"`  //操作是否成功,1成功,0不成功
	Error     string   `json:"error"`   //错误原因
}

//---------------------------------------------------------------------------

