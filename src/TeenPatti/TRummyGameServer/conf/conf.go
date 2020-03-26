package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
	"log"
)

var (
	confPath string
	Conf     = &Config{}
)

type Config struct {
	Server Server
	Desk   DeskData
}

type Server struct {
	IP   string
	Port int
}

type DeskData struct {
	GameStateTime     []int
	FullPlayerNum     int   `toml:"DeskFullplayerNum"`
	MinPlayerNum      int   `toml:"DeskMinplayerNum"`
	Round1GiveUpPoint int32 `toml:"Desk1RoundLosePoint"`
	Round2GiveUpPoint int32 `toml:"Desk2RoundLosePoint"`
	MaxLosePoint      int32 `toml:"DeskMaxLosePoint"`
}

func init() {
	flag.StringVar(&confPath, "conf", "./conf/conf.toml", "-conf path")
	if _, err := toml.DecodeFile(confPath, &Conf); err != nil {
		log.Printf("conf.init() err:%+v", err)
	}
}
