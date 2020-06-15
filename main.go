package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/EasonZhao/tables/gredis"
	"github.com/EasonZhao/tables/logging"
	"github.com/EasonZhao/tables/setting"
)

func main() {
	setting.Setup()
	logging.Setup()
	gredis.Setup(setting.RedisSetting.Host, setting.RedisSetting.Password)
	terminal := make(chan os.Signal)
	signal.Notify(terminal, os.Interrupt, syscall.SIGTERM)
	c := make(chan interface{})
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
