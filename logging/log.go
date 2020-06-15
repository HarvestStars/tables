package logging

import (
	"os"

	"github.com/securityin/auth/pkg/file"
	"github.com/securityin/auth/pkg/setting"
	"github.com/sirupsen/logrus"
)

var (
	f      *os.File
	logger *logrus.Logger
)

// Setup 启动
func Setup() {
	logger = logrus.New()
	var err error
	filePath := getLogFilePath()
	fileName := getLogFileName()
	f, err = file.MustOpen(fileName, filePath)
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Out = f
	logger.SetReportCaller(true)

	if setting.ServerSetting.RunMode == "release" {
		logger.Level = logrus.WarnLevel
	}
}

// GetLogger Logger
func GetLogger() *logrus.Logger {
	return logger
}
