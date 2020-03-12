package game

import (
	"math/rand"
	"time"
	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	"github.com/lonng/nano/scheduler"
	log "github.com/sirupsen/logrus"
	"TeenPatti/TPattiGameServer/protocol"

)

const kickResetBacklog = 8


const (
	versionUpdateMessage  = "你当前的游戏版本过老，请更新客户端，地址: http://fir.im/tand"
)



var defaultManager = NewManager()

type (
	TPManager struct {
		component.Base

		group      *nano.Group       // 广播channel
		players    map[int64]*Player // 所有的玩家

		chKick     chan int64        // 退出队列
		chReset    chan int64        // 重置队列

	}

)

func NewManager() *TPManager {

	return &TPManager{
		group:      nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
		players:    map[int64]*Player{},

		chKick:     make(chan int64, kickResetBacklog),
		chReset:    make(chan int64, kickResetBacklog),
	}
}

func (m *TPManager) AfterInit() {


	session.Lifetime.OnClosed(  func(s *session.Session) {
		m.group.Leave(s)
	})

	// 处理踢出玩家和重置玩家消息(来自http)
	scheduler.NewTimer( time.Second, func()   {

	ctrl:
		for {
			select {
			case uid := <-m.chKick:
				p, ok := defaultManager.player(uid)
				if !ok || p.session == nil {
					logger.Errorf("玩家%d不在线", uid)
				}
				p.session.Close()
				logger.Infof("踢出玩家, UID=%d", uid)

			case uid := <-m.chReset:
				p, ok := defaultManager.player(uid)
				if !ok {
					return
				}
				if p.session != nil {
					logger.Errorf("玩家正在游戏中，不能重置: %d", uid)
					return
				}
				p.desk = nil
				logger.Infof("重置玩家, UID=%d", uid)

			default:
				break ctrl
			}
		}
	})
}

func (m *TPManager) Login(s *session.Session,   req *protocol.LoginToGameServerRequest) error {

	//需要去redis服务器读取 对应的帐号信息
	var (
		Name     string ="aaaaa"
		HeadUrl  string ="main"
		IP       string  = s.RemoteAddr().String()
	)

	logger.Println("Login:Uid:  Id", s.UID() , s.ID() )

	//回复登陆情况
	resp := &protocol.LoginToGameServerResponse{
		Success:  true ,
		Uid:      s.UID(),
		Nickname: Name,
		Sex:      1,
		HeadUrl:  HeadUrl,
	}

	//在此检测版本号
	if req.Version != "Version1.0"  {

		logger.Println("version is too old:", req.Version)

		resp.Success = false
		resp.Error  = versionUpdateMessage

		return s.Response( resp )
	}


	uid := req.Uid      //int64( rand.Int() )  //由服务器产生uid的随机数，到时从redis得到

	err := s.Bind( uid )

	if err != nil  {

		log.Println( "bind error:", uid, err )

		uid = int64( rand.Uint32() )  //由服务器产生uid的随机数，到时从redis得到
	}

	resp.Uid = uid

	if p, ok := m.player(uid); !ok {

		log.Infof("玩家: %d不在线，创建新的玩家 :", uid)

		p = newPlayer(s, uid, Name, HeadUrl, IP, 1 )

		m.setPlayer(uid, p)

	} else {
		log.Infof("玩家: %d已经在线", uid)

		// 移除广播频道
		m.group.Leave(s)

		// 绑定新session
		p.bindSession(s)
	}

	// 添加到广播频道
	m.group.Add(s)

	return s.Response(resp )

}

func (m *TPManager) player(uid int64) (*Player, bool) {


	p, ok := m.players[uid]

	return p, ok
}

func (m *TPManager) setPlayer(uid int64, p *Player) {

	if _, ok := m.players[uid]; ok {
		log.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	m.players[uid] = p
}


func (m *TPManager) sessionCount() int {
	return len(m.players)
}

func (m *TPManager) offline(uid int64) {

	delete(m.players, uid)

	log.Infof("玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
}
