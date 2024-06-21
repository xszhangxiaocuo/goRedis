package tcp

import (
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"goRedis/interface/tcp"
	"goRedis/lib/logger"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type GnetServer struct {
	gnet.BuiltinEventEngine

	eng       gnet.Engine
	addr      string
	multicore bool

	handler tcp.Handler
}

func (es *GnetServer) OnBoot(eng gnet.Engine) gnet.Action {
	es.eng = eng
	log.Printf("echo server with multi-core=%t is listening on %s\n", es.multicore, es.addr)
	return gnet.None
}
func (es *GnetServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("conn accept: %s\n", c.RemoteAddr())
	return nil, gnet.None
}

func (es *GnetServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Printf("conn close: %s\n", c.RemoteAddr())
	c.Close()
	return gnet.None
}

func (es *GnetServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	c.Write(buf)
	return gnet.None
}

func ListenAndServeWithGnet(cfg *Config, handler tcp.Handler) error {
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
