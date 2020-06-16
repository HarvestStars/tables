package setting

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

// App 配置
type App struct {
	Debug           bool
	RuntimeRootPath string

	LogSavePath string
	LogSaveName string
	LogFileExt  string
	TimeFormat  string
}

// AppSetting App 配置
var AppSetting = &App{}

// ElectrumBase 配置
type ElectrumBase struct {
	Host string
}

// ElectrumBaseSetting 数据库配置
var ElectrumBaseSetting = &ElectrumBase{}

// Lavad 配置
type LavadBase struct {
	Host string
	// 可调配的开奖周期间隔，默认2048
	BlocksInSlot int
}

// LavadBaseSetting 数据库配置
var LavadBaseSetting = &LavadBase{}

// Redis 配置
type Redis struct {
	Host     string
	Password string
	// 最大空闲连接数
	MaxIdle int
	// 在给定时间内，允许分配的最大连接数（当为零时，没有限制）
	MaxActive int
	// 在给定时间内将会保持空闲状态，若到达时间限制则关闭连接（当为零时，没有限制）
	IdleTimeout time.Duration
}

// RedisSetting Redis 缓存
var RedisSetting = &Redis{}

var cfg *ini.File

// Setup 初始化各项配置
func Setup() {
	var err error
	cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("Fail to parse 'conf/app.ini': %v", err)
	}

	mapTo("app", AppSetting)
	mapTo("electrum", ElectrumBaseSetting)
	mapTo("redis", RedisSetting)
	mapTo("lavad", LavadBaseSetting)

	RedisSetting.IdleTimeout = RedisSetting.IdleTimeout * time.Second
}

func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo RedisSetting err: %v", err)
	}
}
