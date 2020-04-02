package db

//import (
//	"TeenPatti/TRummyGameServer/conf"
//	"fmt"
//	"github.com/go-redis/redis"
//)
//
//var RedisCon *redis.Client //连接redis
//
//func init() {
//	client := redis.NewClient(&redis.Options{
//		Addr:     conf.Conf.Redis.RedisAddr,
//		Password: conf.Conf.Redis.RedisPassword,
//		DB:       conf.Conf.Redis.RedisDB,
//	})
//
//	_, err := client.Ping().Result()
//	if err != nil {
//		panic("redis连接失败")
//	}
//
//	RedisCon = client
//
//	fmt.Println("Redis连接成功")
//}
