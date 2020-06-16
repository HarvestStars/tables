package logging

import (
	"fmt"
	"time"

	"github.com/HarvestStars/tables/setting"
)

var (
	// LogSavePath 日志路径
	LogSavePath = "runtime/logs/"
	// LogSaveName 日志文件名
	LogSaveName = "log"
	// LogFileExt 日志文件后缀
	LogFileExt = "log"
	// TimeFormat 时间显示格式
	TimeFormat = "20060102"
)

func getLogFilePath() string {
	return fmt.Sprintf("%s", LogSavePath)
}

func getLogFileName() string {
	return fmt.Sprintf("%s%s.%s",
		setting.AppSetting.LogSaveName,
		time.Now().Format(setting.AppSetting.TimeFormat),
		setting.AppSetting.LogFileExt,
	)
}
