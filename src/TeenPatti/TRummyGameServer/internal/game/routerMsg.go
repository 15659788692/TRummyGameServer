package game

//广播
const (
	NoticePlayerJoin   = "Rummy.NoticePlayerJoin"   //广播玩家加入
	NoticeGameState    = "Rummy.NoticeGameState"    //广播游戏状态时间
	NoticeGameStrat    = "Rummy.NoticeGameStrat"    //广播游戏开始
	NoticeGameOperCard = "Rummy.NoticeGameOperCard" //广播玩家的操作
	NoticeGameWin      = "Rummy.NoticeGameWin"      //广播玩家赢牌
	NoticeSettle       = "Rummy.NoticeSettle"       //广播结算
	NoticeGiveUp       = "Rummy.NoticeGiveUp"       //广播玩家放弃
	NoticeEndInfo      = "Rummy.NoticeEndInfo"      //广播结算信息
	NoticePlayerExit   = "Rummy.NoticePlayerExit"   //广播玩家退出
)

//单独推送
const (
	PushSendCard    = "Rummy.PushSendCard"    //发牌
	PushRoundInfo   = "Rummy.PushRoundInfo"   //回合桌子信息通知
	PushOperOutTime = "Rummy.PushOperOutTime" //操作超时
	PushLoseGame    = "Rummy.PushLoseGame"
	PushWinGame     = "Rummy.PushWinGame"
	PushSettle      = "Rummy.PushSettle"
)
