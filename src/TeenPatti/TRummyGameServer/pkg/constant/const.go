package constant

type Behavior int

type DeskStatus int32

const DeskStatusRoute  = "BroadDeskStatus"


const (

	//创建桌子
	DeskStatusCreate        DeskStatus = iota

	//游戏开始倒计时5秒
    DeskStatusReadyStart

	//低注
	DeskStatusReadyLowBet

	//发牌
	DeskStatusReadyFaiPai

	//游戏
	DeskStatusPlaying

	//游戏结术
	DeskStatusRoundOver

	//游戏终/中止
	DeskStatusInterruption

	//已经清洗,即为开桌准备好
	DeskStatusCleaned

	//已销毁
	DeskStatusDestory

)

var stringify = [...]string {

	DeskStatusCreate:       "创建",

	DeskStatusReadyStart:   "开始",
	DeskStatusReadyLowBet:  "低注",
	DeskStatusReadyFaiPai:  "发牌",

	DeskStatusPlaying:      "游戏中",
	DeskStatusRoundOver:    "单局完成",
	DeskStatusInterruption: "游戏终/中止",


	DeskStatusCleaned:      "已清洗",
	DeskStatusDestory:      "已销毁",

}

func (s DeskStatus) String() string {
	return stringify[s]
}
