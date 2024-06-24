package main

import (
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"goRedis/config"
	_ "goRedis/database/cmd"
	"goRedis/lib/logger"
	"goRedis/resp/handler"
	"goRedis/tcp"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
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
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

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

	err := tcp.ListenAndServeWithGnet(&tcp.Options{
		Multicore:               true,              // 启用多核
		LB:                      gnet.RoundRobin,   // 负载均衡策略为轮询
		ReuseAddr:               true,              // 启用 SO_REUSEADDR
		ReusePort:               true,              // 启用 SO_REUSEPORT
		MulticastInterfaceIndex: 0,                 // 默认接口索引为 0
		ReadBufferCap:           65536,             // 读缓冲区大小为 64KB
		WriteBufferCap:          65536,             // 写缓冲区大小为 64KB
		LockOSThread:            false,             // 不锁定 OS 线程
		Ticker:                  true,              // 启用 ticker
		TCPKeepAlive:            30 * time.Second,  // TCP Keep-Alive 设置为 30 秒
		TCPNoDelay:              gnet.TCPNoDelay,   // 禁用 Nagle 算法，即设置为 TCPNoDelay
		LogPath:                 "./gnet.log",      // 日志文件路径
		LogLevel:                logging.InfoLevel, // 日志级别为 Info
		Logger:                  nil,               // 使用默认 logger
		EdgeTriggeredIO:         false,             // 不启用边缘触发 I/O
	},
		&tcp.Config{
			Address:   fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
			Multicore: true,
		},
		handler.NewRESPHandler())

	if err != nil {
		logger.Warn(err)
	}
}
