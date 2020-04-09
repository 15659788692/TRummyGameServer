package game

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var log = logrus.New()

func init() {
	filename := "./log/" + time.Now().Format("20060102150405") + ".log"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		fmt.Println("日志文件创建成功！")
		log.Out = file
	} else {
		fmt.Println("日志文件创建失败！")
		log.Info("Failed to log to file")
	}
}
