package game

const (
	NoticePlayerJoin   = "Rummy.NoticePlayerJoin"   //广播玩家加入
	NoticeGameState    = "Rummy.NoticeGameState"    //广播游戏状态时间
	NoticeGameStrat    = "Rummy.NoticeGameStrat"    //广播游戏开始
	NoticeGameOperCard = "Rummy.NoticeGameOperCard" //广播玩家的操作
	NoticeGameWin      = "Rummy.NoticeGameWin"      //广播玩家赢牌
)

const ( //单独通知
	NoticeDeskInfo = "Rummy.NoticeDeskInfo" //玩家桌子信息
	NoticeSendCard = "Rummy.NoticeSendCard" //发牌
	NoticePlayEnd  = "Rummy.NoticePlayEnd"
)
