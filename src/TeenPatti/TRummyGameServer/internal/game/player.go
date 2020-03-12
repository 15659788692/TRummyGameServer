package game

import (
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
)

type Loser struct {
	uid   int64
	score int
}

type Player struct {
	uid        int64  // 用户ID
	seatPos    int    //座位号（系统从0开始，最多4,  每张桌只能有5个玩家)
	head       string // 头像地址
	name       string // 玩家名字
	ip         string // ip地址
	sex        int    // 性别
	level      int    //玩家的等级
	starNum    int    //星的个数
	isJoin     bool   //是否进入圈子游戏,主要用于当进入游戏时
	sitdown    bool   //是否坐下
	disconnect bool   //是否掉线

	session *session.Session //玩家对应的网络通道
	desk    *Desk            //玩家的桌子
	logger  *log.Entry       //日志文件

	win       int64 //在此桌的胜负
	isBanker  bool  //是否是庄家
	deposit   bool  //是否托管
	settle    bool  //是否结算
	showed    bool  //是否已show
	HandCards []GCard
	Timeout   int32 //连续超时次数
}

func newPlayer(s *session.Session, uid int64, nicename, head, ip string, sex int) *Player {

	p := &Player{

		uid:     uid,
		seatPos: -1, //还未入座
		name:    nicename,
		head:    head,

		ip:         ip,
		sex:        sex,
		disconnect: false,
		deposit:    false,
		Timeout:    0,
		settle:     false,
		logger:     log.WithField("player", uid),
	}

	//绑定对应的session
	p.bindSession(s)

	return p
}

//初始化玩家
func (p *Player) InitPlayer() {
	p.win = 0
	p.isBanker = false
	p.deposit = false
	p.settle = false
	p.showed = false
	p.HandCards = []GCard{}
	p.Timeout = 0
}

func (p *Player) bindSession(s *session.Session) {

	p.session = s
	p.session.Set(kCurPlayer, p)

}

func (p *Player) removeSession() {
	p.session.Remove(kCurPlayer)
	p.session = nil
}

func (p *Player) setDesk(d *Desk, turn int) {

	if d == nil {
		p.logger.Error("桌号为空")
		return
	}

	p.desk = d
	p.seatPos = turn
	p.logger = log.WithFields(log.Fields{"deskno": p.desk.roomNo, "player": p.uid})

}

func (p *Player) Uid() int64 {
	return p.uid
}

func (p *Player) setIp(ip string) {
	p.ip = ip
}

//设定座位号
func (p *Player) SetSeatPos(seatPos int) {

	p.seatPos = seatPos
	if seatPos >= 0 {
		p.sitdown = true
	} else {
		p.sitdown = false
	}
}

//读取座位号
func (p *Player) GetSeatPos() int {

	return p.seatPos
}

//设定椅子坐下状态
func (p *Player) SetSitdown(bsitdown bool) {

	p.sitdown = bsitdown
}

//读取当前的椅子状态
func (p *Player) GetSitdown() bool {

	return p.sitdown
}

// 断线重连后，同步牌桌数据
//	TODO: 断线重连，已和牌玩家显示不正常
func (p *Player) syncDeskData() error {

	return nil
}

//删除手中的牌
func (p *Player) DelHandCard(card int32) bool {
	//检测玩家手中是否有手牌
	if len(p.HandCards) <= 1 {
		return false
	}
	//检测手牌中是否有这张牌
	for k, v := range p.HandCards {
		if v.Card == card {
			p.HandCards = append(p.HandCards[:k], p.HandCards[k+1:]...)
			return true
		}
	}
	return false
}
