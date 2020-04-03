package game

import (
	"TeenPatti/TRummyGameServer/conf"
)

var (
	GetPlayerMsgFromRedis = conf.Conf.Server.IP + "/v1/user/query"   //获取用户信息
	GetFriendsMsg         = conf.Conf.Server.IP + "/v1/user/friends" //获取好友信息
	DeleteFriends         = conf.Conf.Server.IP + "/v1/user/friends" //删除好友信息
)

//广播
const (
	NoticePlayerState  = "Rummy.NoticePlayerState"  //广播玩家状态
	NoticeGameState    = "Rummy.NoticeGameState"    //广播游戏状态时间
	NoticeGameOperCard = "Rummy.NoticeGameOperCard" //广播玩家的操作
	NoticeGameWin      = "Rummy.NoticeGameWin"      //广播玩家赢牌
	NoticeSettle       = "Rummy.NoticeSettle"       //广播结算
	NoticeGiveUp       = "Rummy.NoticeGiveUp"       //广播玩家放弃
	NoticeEndInfo      = "Rummy.NoticeEndInfo"      //广播结算信息
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
