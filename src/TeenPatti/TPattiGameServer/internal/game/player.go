package game

import (
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
	"time"
)

type Loser struct {
	uid   int64
	score int
}

type Player struct {

	uid       int64       // 用户ID

	seatPos    int    //座位号（系统从0开始，最多4,  每张桌只能有5个玩家)

	head string      // 头像地址
	name string      // 玩家名字

	ip   string      // ip地址
	sex  int         // 性别

	level   int      //玩家的等级
	starNum int      //星的个数


	betScore   int   //玩家的下注
	win        int   //在此桌的胜负

	bet      int     //投注的额度

	betstarttime        time.Time   //轮到自已下注的开始时间
	joinTime            time.Time   //加入此桌的时间


	isJoin      bool      //是否进入圈子游戏,主要用于当进入游戏时
	isCanBet    bool      //是轮到自已下注吗
	isBanker    bool      //是否是庄家


	sitdown     bool      //是否坐下

	packed      bool      //是否已pack
	showed      bool      //是否已show
	blinded     bool      //是否显示blind
	seen        bool      //是否已看牌


	pair     [3]byte      //发给的3张扑克牌

	betSignal  chan int   //投注的信号量,当有投注时


	session    *session.Session  //玩家对应的网络通道
	desk       *Desk             //玩家的桌子
	logger     *log.Entry        //日志文件
}



func newPlayer(s *session.Session, uid int64, nicename, head, ip string,   sex int) *Player {

	p := &Player{

		uid:   uid,
		name:  nicename,
		head:  head,

		ip:    ip,
		sex:   sex,
		betScore: 0,

		betSignal:make( chan int ),

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

func (p *Player) Uid() int64 {
	return p.uid
}


func (p *Player) setIp(ip string) {
	p.ip = ip
}

//设定座位号
func ( p* Player ) SetSeatPos(  seatPos int  ) {

	p.seatPos = seatPos
}

//读取座位号
func ( p *Player ) GetSeatPos( ) int {

	return p.seatPos
}

//读取投注的通道信号
func ( p *Player ) GetBetSignal() ( chan int ){

	return p.betSignal
}

//设定投注的通道信号
func ( p *Player ) SetBetSignal( betCount int ) {

	p.betSignal<- betCount
}

//设定椅子坐下状态
func (p *Player ) SetSitdown( bsitdown bool  ) {

	p.sitdown = bsitdown
}

//读取当前的椅子状态
func (p *Player ) GetSitdown( ) bool {

	return p.sitdown
}


//是否可下注标志
func ( p *Player ) GetIsCanBet() bool {

	return p.isCanBet
}

//设定下注标志
func ( p *Player ) SetIsCanBet(  isBet bool  ) {

	p.isCanBet = isBet
}


//检测玩家是否在下注
func ( p *Player) IsBetting() bool {

	if p.desk == nil  {
		return false
	}

	//检测桌子是否可投注
	bBet :=p.desk.DeskIsCanBet()

	if bBet == false  {
		return true
	}

	//检测当前玩家是可否投
	if p.GetIsCanBet() == false {
		return false
	}

	//若是没有坐下也不可以投注
	if p.GetSitdown() == false {

		return false
	}


	return true
}

//投注,已返回投注额度
func ( p *Player ) PlayBet( betCount int  )( int, error ) {

	   p.betScore += betCount

       return  p.betScore , nil
}

// 断线重连后，同步牌桌数据
// TODO: 断线重连，已和牌玩家显示不正常
func (p *Player) syncDeskData() error {

	return nil
}



