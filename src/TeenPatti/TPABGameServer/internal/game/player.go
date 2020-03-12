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

	uid  int64  // 用户ID
	head string // 头像地址
	name string // 玩家名字

	ip   string // ip地址
	sex  int    // 性别

	// 玩家数据
	session *session.Session

	desk     *Desk //当前桌


	score int   //经过n局后,当前玩家余下的分值数,默认为1000

	logger *log.Entry // 日志
}

func newPlayer(s *session.Session, uid int64, name, head, ip string, sex int) *Player {

	p := &Player{

		uid:   uid,
		name:  name,
		head:  head,

		ip:    ip,
		sex:   sex,
		score: 1000,

		logger: log.WithField("player", uid),

	}

	//绑定对应的session
	p.bindSession(s)

	return p
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
	p.logger = log.WithFields(log.Fields{"deskno": p.desk.roomNo, "player": p.uid})

}

func (p *Player) setIp(ip string) {
	p.ip = ip
}


func (p *Player) Uid() int64 {
	return p.uid
}


// 断线重连后，同步牌桌数据
// TODO: 断线重连，已和牌玩家显示不正常
func (p *Player) syncDeskData() error {

	return nil

}
