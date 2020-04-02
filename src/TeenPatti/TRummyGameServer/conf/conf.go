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
	Server ServerConfig `toml:"server"`
	Desk   DeskData     `toml:"desk"`
	Redis  RedisConfig  `toml:"redis"`
	DB     DataBase     `toml:"db"`
}

//redis
type RedisConfig struct {
	RedisAddr     string `toml:"redisAddr"`
	RedisPassword string `toml:"redisPassword"`
	RedisDB       int    `toml:"redisDB"`
}

//数据库
type DataBase struct {
	HOST     string `toml:"host"`
	PORT     int32  `toml:"post"`
	DBNAME   string `toml:"dbName"`
	USERNAME string `toml:"userName"`
	PASSWORD string `toml:"password"`
	NETWORK  string `toml:"network"`
}

type ServerConfig struct {
	IP   string `toml:"ip"`
	Port int    `toml:"port"`
}

type DeskData struct {
	GameStateTime     []int
	FullPlayerNum     int   `toml:"DeskFullplayerNum"`
	MinPlayerNum      int   `toml:"DeskMinplayerNum"`
	Round1GiveUpPoint int32 `toml:"Desk1RoundLosePoint"`
	Round2GiveUpPoint int32 `toml:"Desk2RoundLosePoint"`
	MaxLosePoint      int32 `toml:"DeskMaxLosePoint"`
	//测试
	FirstPlayerCards []int32 `toml:"FirstPlayerCards"`
	WildCard         int32   `toml:"WildCard"`
}

func init() {
	var err error
	flag.StringVar(&confPath, "conf", "./src/TeenPatti/TRummyGameServer/conf/conf.toml", "-conf path")
	if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
		log.Printf("conf.init() err:%+v", err)
		flag.StringVar(&confPath, "conf2", "./conf/conf.toml", "-conf path")
		if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
			log.Printf("conf.init() err:%+v", err)
		}
	}

}
