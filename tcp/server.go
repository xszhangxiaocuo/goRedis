package tcp

import (
	"context"
	"fmt"
	"goRedis/interface/tcp"
	"goRedis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string //tcp监听端口
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	//syscall.SIGHUP：通常表示终端断开或者控制进程结束，常用于通知守护进程重新读取配置文件。
	//syscall.SIGQUIT：通常表示用户请求退出并生成核心转储（core dump），用于调试。
	//syscall.SIGTERM：是一个终止信号，通常用于请求程序正常退出。
	//syscall.SIGINT：通常由用户通过控制台（如按下 Ctrl+C）发送，请求中断程序。
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		//判断接收的信号类型
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{} //发送关闭信号
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("tcp start listen at:%s", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() { //这个协程在用户强制关闭程序或系统kill掉程序进程时触发，对连接进行关闭
		<-closeChan //没有收到数据就会一直阻塞
		logger.Info("shutting down...")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup //超时控制
	for {
		logger.Info("waiting link...")
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Wait()
			}()
			handler.Handler(ctx, conn)
		}()
	}
	waitDone.Wait()
}
