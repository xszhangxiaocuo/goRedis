package main

import (
	"fmt"
	"goRedis/config"
	"goRedis/lib/logger"
	"goRedis/tcp"
	"os"
)

const configFile string = "redis.conf"

var defaultConfig = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

// 判断文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func main() {
	logger.Setup(&logger.Settings{
		Path:       "log",
		Name:       "go_redis",
		Ext:        "log",
		TimeFormat: "2006-01-02 15:04:05",
	})

	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultConfig
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		},
		tcp.NewEchoHandler())

	if err != nil {
		logger.Warn(err)
	}
}
