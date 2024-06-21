package tcp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"goRedis/interface/tcp"
	"goRedis/resp/connection"
	"log"
	"sync"
	"time"
)

type Options struct {
	// Multicore indicates whether the server will be effectively created with multi-cores, if so,
	// then you must take care with synchronizing memory between all event callbacks, otherwise,
	// it will run the server with single thread. The number of threads in the server will be automatically
	// assigned to the value of logical CPUs usable by the current process.
	Multicore bool

	// LockOSThread is used to determine whether each I/O event-loop is associated to an OS thread, it is useful when you
	// need some kind of mechanisms like thread local storage, or invoke certain C libraries (such as graphics lib: GLib)
	// that require thread-level manipulation via cgo, or want all I/O event-loops to actually run in parallel for a
	// potential higher performance.
	LockOSThread bool

	// ReadBufferCap is the maximum number of bytes that can be read from the client when the readable event comes.
	// The default value is 64KB, it can be reduced to avoid starving subsequent client connections.
	//
	// Note that ReadBufferCap will be always converted to the least power of two integer value greater than
	// or equal to its real amount.
	ReadBufferCap int

	// LB represents the load-balancing algorithm used when assigning new connections.
	LB gnet.LoadBalancing

	// NumEventLoop is set up to start the given number of event-loop goroutine.
	// Note: Setting up NumEventLoop will override Multicore.
	NumEventLoop int

	// ReusePort indicates whether to set up the SO_REUSEPORT socket option.
	ReusePort bool

	// Ticker indicates whether the ticker has been set up.
	Ticker bool

	// TCPKeepAlive sets up a duration for (SO_KEEPALIVE) socket option.
	TCPKeepAlive time.Duration

	// TCPNoDelay controls whether the operating system should delay
	// packet transmission in hopes of sending fewer packets (Nagle's algorithm).
	//
	// The default is true (no delay), meaning that data is sent
	// as soon as possible after a Write.
	TCPNoDelay gnet.TCPSocketOpt

	// SocketRecvBuffer sets the maximum socket receive buffer in bytes.
	SocketRecvBuffer int

	// SocketSendBuffer sets the maximum socket send buffer in bytes.
	SocketSendBuffer int
}

type Command struct {
	Args [][]byte // 命令参数
	Raw  []byte   // 原始命令
}

type connBuffer struct {
	buf     bytes.Buffer
	command []Command
}

type GnetServer struct {
	*gnet.BuiltinEventEngine
	eng        gnet.Engine
	ActiveConn sync.Map // 当前存活的连接
	addr       string
	multicore  bool

	handler   tcp.Handler
	closeChan chan struct{}
}

func (gs *GnetServer) OnBoot(eng gnet.Engine) gnet.Action {
	gs.eng = eng
	log.Printf("go_redis server with multi-core=%t is listening on %s\n", gs.multicore, gs.addr)
	return gnet.None
}
func (gs *GnetServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	//log.Printf("conn accept: %s\n", c.RemoteAddr())
	gs.ActiveConn.Store(c, connection.NewRESPConn(c))
	return nil, gnet.None
}

func (gs *GnetServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	//log.Printf("conn close: %s\n", c.RemoteAddr())
	gs.ActiveConn.Delete(c)
	return gnet.Close
}

func (gs *GnetServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := context.Background()
	val, ok := gs.ActiveConn.Load(c)
	client := val.(*connection.RESPConn)
	if !ok {
		//logging.Infof("连接已经关闭：%s", c.RemoteAddr())
		return gnet.None
	}
	gs.handler.Handler(ctx, client)
	return gnet.None
}

func ListenAndServeWithGnet(options *Options, cfg *Config, handler tcp.Handler) error {
	//closeChan := make(chan struct{})
	//sigChan := make(chan os.Signal)
	////syscall.SIGHUP：通常表示终端断开或者控制进程结束，常用于通知守护进程重新读取配置文件。
	////syscall.SIGQUIT：通常表示用户请求退出并生成核心转储（core dump），用于调试。
	////syscall.SIGTERM：是一个终止信号，通常用于请求程序正常退出。
	////syscall.SIGINT：通常由用户通过控制台（如按下 Ctrl+C）发送，请求中断程序。
	//signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	//go func() {
	//	sig := <-sigChan
	//	switch sig {
	//	//判断接收的信号类型
	//	case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
	//		closeChan <- struct{}{} //发送关闭信号
	//	}
	//}()

	gs := &GnetServer{
		addr:      fmt.Sprintf("tcp://%s", cfg.Address),
		multicore: cfg.Multicore,
		handler:   handler,
		//closeChan: closeChan,
	}
	//go func() {
	//	<-gs.closeChan
	//	gs.OnShutdown(gs.eng)
	//	gs.handler.Close()
	//}()
	logging.Infof("gnet server is starting at %s", cfg.Address)

	serveOptions := gnet.Options{
		Multicore:        options.Multicore,
		LockOSThread:     options.LockOSThread,
		ReadBufferCap:    options.ReadBufferCap,
		LB:               options.LB,
		NumEventLoop:     options.NumEventLoop,
		ReusePort:        options.ReusePort,
		Ticker:           options.Ticker,
		TCPKeepAlive:     options.TCPKeepAlive,
		TCPNoDelay:       options.TCPNoDelay,
		SocketRecvBuffer: options.SocketRecvBuffer,
		SocketSendBuffer: options.SocketSendBuffer,
	}
	return gnet.Run(gs, gs.addr, gnet.WithOptions(serveOptions))
}
