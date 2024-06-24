package tcp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"goRedis/interface/tcp"
	"goRedis/resp/connection"
	"log"
	"sync"
	"time"
)

type Options struct {
	// Multicore indicates whether the engine will be effectively created with multi-cores, if so,
	// then you must take care with synchronizing memory between all event callbacks, otherwise,
	// it will run the engine with single thread. The number of threads in the engine will be automatically
	// assigned to the value of logical CPUs usable by the current process.
	Multicore bool

	// NumEventLoop is set up to start the given number of event-loop goroutine.
	// Note: Setting up NumEventLoop will override Multicore.
	NumEventLoop int

	// LB represents the load-balancing algorithm used when assigning new connections.
	LB gnet.LoadBalancing

	// ReuseAddr indicates whether to set up the SO_REUSEADDR socket option.
	ReuseAddr bool

	// ReusePort indicates whether to set up the SO_REUSEPORT socket option.
	ReusePort bool

	// MulticastInterfaceIndex is the index of the interface name where the multicast UDP addresses will be bound to.
	MulticastInterfaceIndex int

	// ============================= Options for both server-side and client-side =============================

	// ReadBufferCap is the maximum number of bytes that can be read from the remote when the readable event comes.
	// The default value is 64KB, it can either be reduced to avoid starving the subsequent connections or increased
	// to read more data from a socket.
	//
	// Note that ReadBufferCap will always be converted to the least power of two integer value greater than
	// or equal to its real amount.
	ReadBufferCap int

	// WriteBufferCap is the maximum number of bytes that a static outbound buffer can hold,
	// if the data exceeds this value, the overflow will be stored in the elastic linked list buffer.
	// The default value is 64KB.
	//
	// Note that WriteBufferCap will always be converted to the least power of two integer value greater than
	// or equal to its real amount.
	WriteBufferCap int

	// LockOSThread is used to determine whether each I/O event-loop is associated to an OS thread, it is useful when you
	// need some kind of mechanisms like thread local storage, or invoke certain C libraries (such as graphics lib: GLib)
	// that require thread-level manipulation via cgo, or want all I/O event-loops to actually run in parallel for a
	// potential higher performance.
	LockOSThread bool

	// Ticker indicates whether the ticker has been set up.
	Ticker bool

	// TCPKeepAlive sets up a duration for (SO_KEEPALIVE) socket option.
	TCPKeepAlive time.Duration

	// TCPNoDelay controls whether the operating system should delay
	// packet transmission in hopes of sending fewer packets (Nagle's algorithm).
	//
	// The default is true (no delay), meaning that data is sent
	// as soon as possible after a write operation.
	TCPNoDelay gnet.TCPSocketOpt

	// SocketRecvBuffer sets the maximum socket receive buffer in bytes.
	SocketRecvBuffer int

	// SocketSendBuffer sets the maximum socket send buffer in bytes.
	SocketSendBuffer int

	// LogPath the local path where logs will be written, this is the easiest way to set up logging,
	// gnet instantiates a default uber-go/zap logger with this given log path, you are also allowed to employ
	// you own logger during the lifetime by implementing the following log.Logger interface.
	//
	// Note that this option can be overridden by the option Logger.
	LogPath string

	// LogLevel indicates the logging level, it should be used along with LogPath.
	LogLevel logging.Level

	// Logger is the customized logger for logging info, if it is not set,
	// then gnet will use the default logger powered by go.uber.org/zap.
	Logger logging.Logger

	// EdgeTriggeredIO enables the edge-triggered I/O for the underlying epoll/kqueue event-loop.
	// Don't enable it unless you are 100% sure what you are doing.
	// Note that this option is only available for stream-oriented protocol.
	EdgeTriggeredIO bool
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
	pool       *goroutine.Pool
	handler    tcp.Handler
	closeChan  chan struct{}
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
	var result []byte
	action = gnet.None
	ctx := context.Background()
	val, ok := gs.ActiveConn.Load(c)
	client := val.(*connection.RESPConn)
	if !ok {
		//logging.Infof("连接已经关闭：%s", c.RemoteAddr())
		return
	}
	data, _ := c.Next(-1)
	_ = gs.pool.Submit(func() {
		result, action = gs.handler.Handler(ctx, client, data)
		c.Write(result)
	})

	return
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
	p := goroutine.Default()
	defer p.Release()
	gs := &GnetServer{
		addr:      fmt.Sprintf("tcp://%s", cfg.Address),
		multicore: cfg.Multicore,
		pool:      p,
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
