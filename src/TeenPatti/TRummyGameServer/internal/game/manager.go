package game

import (
	"TeenPatti/TRummyGameServer/db"
	"TeenPatti/TRummyGameServer/protocol"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"
)

const kickResetBacklog = 8

const (
	versionUpdateMessage = "你当前的游戏版本过老，请更新客户端，地址: http://fir.im/tand"
)

var defaultManager = NewManager()

type (
	TRManager struct {
		component.Base

		group   *nano.Group       // 广播channel
		players map[int64]*Player // 所有的玩家

		chKick  chan int64 // 退出队列
		chReset chan int64 // 重置队列

	}
)

func NewManager() *TRManager {

	return &TRManager{
		group:   nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
		players: map[int64]*Player{},

		chKick:  make(chan int64, kickResetBacklog),
		chReset: make(chan int64, kickResetBacklog),
	}
}

func (m *TRManager) AfterInit() {

	session.Lifetime.OnClosed(func(s *session.Session) {
		m.group.Leave(s)
	})

	// 处理踢出玩家和重置玩家消息(来自http)
	scheduler.NewTimer(time.Second, func() {

	ctrl:
		for {
			select {
			case uid := <-m.chKick:
				p, ok := defaultManager.player(uid)
				if !ok || p.session == nil {
					logger.Errorf("玩家%d不在线", uid)
				}
				if p.desk != nil {
					p.session.Close()
				}
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

func (m *TRManager) Login(s *session.Session, req *protocol.LoginToGameServerRequest) error {

	IsLogin := CheckPlayerIsLoginFromRedis(strconv.Itoa(int(req.Uid)), req.Token)
	if !IsLogin {
		return s.Response(&protocol.LoginToGameServerResponse{
			Success: false,
			Error:   "Token or UID is error！",
		})
	}
	//获取用户基础信息
	playMsg := GetPlayerMsgFromFaceBook(req.Token)
	log.Println(playMsg)
	if playMsg == nil {
		return s.Response(&protocol.LoginToGameServerResponse{
			Success: false,
			Error:   "get user message failed!",
		})
	}
	log.Println("玩家的信息:", playMsg)

	logger.Println("Login:Uid:  Id", s.UID(), s.ID())
	//回复登陆情况
	resp := &protocol.LoginToGameServerResponse{
		Success:  true,
		Uid:      req.Uid,
		Nickname: playMsg.Name,
		Sex:      1,
		HeadUrl:  playMsg.Avatar,
	}

	//在此检测版本号
	if req.Version != "Version1.0" {
		logger.Println("version is too old:", req.Version)
		resp.Success = false
		resp.Error = versionUpdateMessage
		return s.Response(resp)
	}

	uid := req.Uid //int64( rand.Int() )  //由服务器产生uid的随机数，到时从redis得到

	err := s.Bind(uid)

	if err != nil {

		log.Println("bind error:", uid, err)

		uid = req.Uid //由服务器产生uid的随机数，到时从redis得到
	}

	resp.Uid = uid

	if p, ok := m.player(uid); !ok {

		log.Infof("玩家: %d不在线，创建新的玩家 :", uid)

		p = newPlayer(s, uid, playMsg.Name, playMsg.Avatar, "IP", 1)

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

	return s.Response(resp)

}

func (m *TRManager) player(uid int64) (*Player, bool) {

	p, ok := m.players[uid]

	return p, ok
}

func (m *TRManager) setPlayer(uid int64, p *Player) {

	if _, ok := m.players[uid]; ok {
		log.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	m.players[uid] = p
}

func (m *TRManager) sessionCount() int {
	return len(m.players)
}

func (m *TRManager) offline(uid int64) {

	delete(m.players, uid)

	log.Infof("玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
}

func CheckPlayerIsLoginFromRedis(id string, token string) bool {
	val, err := db.RedisCon.Get(fmt.Sprintf("session:%v", id)).Result()
	if err != nil {
		log.Println("获取token错误")
		return false
	}

	if token == val {
		return true
	}

	log.Println("token不相等", token)
	log.Println(val)

	return false
}

//获取用户信息从facebook
func GetPlayerMsgFromFaceBook(token string) *protocol.FaceBookGetPlayerMsg {
	httpreq, err1 := http.NewRequest(http.MethodGet, GetPlayerMsgFromRedis, nil)
	if err1 != nil {
		log.Println("获取用户信息请求创建失败！")
		return nil
	}

	httpreq.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	resp, err2 := http.DefaultClient.Do(httpreq)
	if err2 != nil && resp.StatusCode == 200 {
		log.Println("获取用户信息请求失败", err2)
		return nil
	}
	defer resp.Body.Close()

	respByte, _ := ioutil.ReadAll(resp.Body)

	var data protocol.FaceBookGetPlayerMsgData
	_ = json.Unmarshal(respByte, &data)
	if data.Code != 0 {
		return nil
	}

	return &data.Data
}
