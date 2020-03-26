package game

import (
	"fmt"
	"math/rand"

	// "net/http"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"
)
import "TeenPatti/TRummyGameServer/conf"

var (
	version = ""            // 游戏版本
	consume = map[int]int{} // 房卡消耗配置

	forceUpdate = false

	logger = log.WithField("component", "game")
)

// Startup 初始化游戏服务器
func Startup() {

	rand.Seed(time.Now().Unix())

	//	version = viper.GetString("update.version")
	//	heartbeat := viper.GetInt("core.heartbeat")

	heartbeat := 10

	if heartbeat < 5 {
		heartbeat = 5
	}

	//	logger.Infof("当前游戏服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("game service starup")

	// register game handler
	comps := &component.Components{}
	comps.Register(defaultManager)
	comps.Register(defaultDeskManager)

	// 加密管道
	//c := newCrypto()
	//pip := pipeline.New()
	//pip.Inbound().PushBack(c.inbound)
	//pip.Outbound().PushBack(c.outbound)
	//	addr := fmt.Sprintf(":%d", viper.GetInt("game-server.port"))

	addr := fmt.Sprintf(":%d", conf.Conf.Server.Port)

	/*nano.Listen(addr,
		nano.WithHeartbeatInterval(time.Duration(heartbeat)*time.Second),
		nano.WithLogger(log.WithField("component", "ABGame")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
		nano.WithIsWebsocket(true ),
		nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }),
		nano.WithDebugMode(),

	)*/

	nano.Listen(addr,
		nano.WithHeartbeatInterval(time.Duration(heartbeat)*time.Second),
		nano.WithLogger(log.WithField("component", "RummyGame")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
		nano.WithIsWebsocket(true),
		// nano.WithCheckOriginFunc(func(_ *http.Request) bool { return true }), //
		nano.WithDebugMode(), //debug模式
	)
}
