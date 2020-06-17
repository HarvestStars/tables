package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/HarvestStars/tables/gredis"
	"github.com/HarvestStars/tables/logging"
	"github.com/HarvestStars/tables/setting"
)

func main() {
	setting.Setup()
	logging.Setup()
	gredis.Setup(setting.RedisSetting.Host, setting.RedisSetting.Password)
	lavadClient = &http.Client{}
	terminal := make(chan os.Signal)
	signal.Notify(terminal, os.Interrupt, syscall.SIGTERM)
	c := make(chan interface{})
	logging.GetLogger().Info("version 1.0")
	go run(c)
	for {
		select {
		case <-terminal:
			fmt.Println("\r- Ctrl+C pressed in Terminal")
			c <- 0
		case <-c:
			fmt.Println("exit")
			return
		}
	}
}
