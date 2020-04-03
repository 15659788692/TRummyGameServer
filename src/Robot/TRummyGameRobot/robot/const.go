package robot

//广播
const (
	NoticePlayerJoin   = "Rummy.NoticePlayerJoin"   //广播玩家加入
	NoticeGameState    = "Rummy.NoticeGameState"    //广播游戏状态时间
	NoticeGameOperCard = "Rummy.NoticeGameOperCard" //广播玩家的操作
	NoticeGameWin      = "Rummy.NoticeGameWin"      //广播玩家赢牌
	NoticeSettle       = "Rummy.NoticeSettle"       //广播结算
	NoticeGiveUp       = "Rummy.NoticeGiveUp"       //广播玩家放弃
	NoticeEndInfo      = "Rummy.NoticeEndInfo"      //广播结算信息
	NoticePlayerExit   = "Rummy.NoticePlayerExit"   //广播玩家退出
)

//单独推送
const (
	PushGameStrat   = "Rummy.PushGameStrat"   //广播游戏开始
	PushSendCard    = "Rummy.PushSendCard"    //发牌
	PushRoundInfo   = "Rummy.PushRoundInfo"   //回合桌子信息通知
	PushOperOutTime = "Rummy.PushOperOutTime" //操作超时
	PushLoseGame    = "Rummy.PushLoseGame"
	PushWinGame     = "Rummy.PushWinGame"
	PushSettle      = "Rummy.PushSettle"
)

//请求
const (
	ReqLogin        = "TRManager.Login"
	ReqDeskJoinDesk = "TRDeskManager.JoinDesk"
	ReqDeskOperCard = "TRDeskManager.OperCard"
	ReqDeskShowCard = "TRDeskManager.ShowCard"
	ReqDeskSettle   = "TRDeskManager.Settle"
	ReqDeskGiveUp   = "TRDeskManager.GiveUp"
)

const ( //游戏状态
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
