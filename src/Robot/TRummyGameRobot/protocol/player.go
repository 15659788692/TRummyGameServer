package protocol

type NoMsg struct{}

//Do:玩家信息
type PlayerMsg struct {
	UserName     string  //用户名
	Gold         float32 //金币
	HeadPortrait string  //头像
	Level        int8    //等级
	Exp          int16   //经验
	UserAndar    float32 //玩家在A区下注金额
	UserBahar    float32 //玩家在B区下注金额
	Seat         int8    //座位
}

type PlayerJoinRoomRequest struct {
	UserID string
}
