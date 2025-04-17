package client

import (
	"errors"
	"goRedis/interface/resp"
	"goRedis/lib/logger"
	"goRedis/lib/sync/wait"
	"goRedis/resp/parser"
	"goRedis/resp/reply"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	created = iota // 已创建
	running        // 运行中
	closed         // 已关闭
)

// Client 是一个支持管道模式的 redis 客户端
type Client struct {
	conn        net.Conn      // TCP 连接
	pendingReqs chan *request // 待发送的请求
	waitingReqs chan *request // 等待响应的请求
	ticker      *time.Ticker  // 心跳定时器
	addr        string        // 远程地址

	status  int32           // 客户端状态
	working *sync.WaitGroup // 用于统计未完成的请求（包含待发送和等待响应的）

	tickerHook func() // 心跳钩子函数
}

// request 表示发送给 redis 服务端的一条消息
type request struct {
	id        uint64     // 请求 ID
	args      [][]byte   // 请求参数
	reply     resp.Reply // 服务端响应
	heartbeat bool       // 是否是心跳请求
	waiting   *wait.Wait // 等待器
	err       error      // 错误信息
}

const (
	chanSize = 256             // channel 缓存大小
	maxWait  = 3 * time.Second // 最大等待时间
)

// MakeClient 创建一个新的客户端实例
func MakeClient(addr string, tickerHook func()) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		addr:        addr,
		conn:        conn,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
		working:     &sync.WaitGroup{},
		tickerHook:  tickerHook,
	}, nil
}

// RemoteAddress 返回远程地址
func (client *Client) RemoteAddress() string {
	return client.addr
}

// Start 启动异步协程
func (client *Client) Start() {
	client.ticker = time.NewTicker(10 * time.Second)
	go client.handleWrite() // 处理写请求
	go client.handleRead()  // 处理读响应
	go client.heartbeat()   // 启动心跳机制
	atomic.StoreInt32(&client.status, running)
}

// Close 停止异步协程并关闭连接
func (client *Client) Close() {
	atomic.StoreInt32(&client.status, closed)
	client.ticker.Stop()
	close(client.pendingReqs)

	client.working.Wait()

	_ = client.conn.Close()
	close(client.waitingReqs)
}

// reconnect 重连 redis 服务端
func (client *Client) reconnect() {
	logger.Info("reconnect with: " + client.addr)
	_ = client.conn.Close() // 忽略重复关闭的错误

	var conn net.Conn
	for i := 0; i < 3; i++ {
		var err error
		conn, err = net.Dial("tcp", client.addr)
		if err != nil {
			logger.Error("reconnect error: " + err.Error())
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}
	if conn == nil { // 达到最大重试次数，关闭客户端
		if client.tickerHook != nil {
			client.tickerHook()
		}
		client.Close()
		return
	}
	client.conn = conn

	// 通知等待响应的请求失败
	close(client.waitingReqs)
	for req := range client.waitingReqs {
		req.err = errors.New("connection closed")
		req.waiting.Done()
	}
	client.waitingReqs = make(chan *request, chanSize)
	// 重新启动读协程
	go client.handleRead()
}

// heartbeat 心跳检测，定时发送 PING
func (client *Client) heartbeat() {
	for range client.ticker.C {
		client.doHeartbeat()
	}
}

// handleWrite 处理写请求，将请求写入 TCP 连接
func (client *Client) handleWrite() {
	for req := range client.pendingReqs {
		client.doRequest(req)
	}
}

// Send 发送一条请求到 redis 服务端
func (client *Client) Send(args [][]byte) resp.Reply {
	if atomic.LoadInt32(&client.status) != running {
		return reply.NewStandardErrReply("client closed")
	}
	req := &request{
		args:      args,
		heartbeat: false,
		waiting:   &wait.Wait{},
	}
	req.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- req
	timeout := req.waiting.WaitWithTimeout(maxWait)
	if timeout {
		return reply.NewStandardErrReply("server time out")
	}
	if req.err != nil {
		return reply.NewStandardErrReply("request failed " + req.err.Error())
	}
	return req.reply
}

// doHeartbeat 发送心跳请求（PING）
func (client *Client) doHeartbeat() {
	request := &request{
		args:      [][]byte{[]byte("PING")},
		heartbeat: true,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- request
	request.waiting.WaitWithTimeout(maxWait)
}

// doRequest 执行请求，将请求数据写入连接
func (client *Client) doRequest(req *request) {
	if req == nil || len(req.args) == 0 {
		return
	}
	re := reply.NewMultiBulkReply(req.args)
	bytes := re.ToBytes()
	var err error
	for i := 0; i < 3; i++ {
		_, err = client.conn.Write(bytes)
		if err == nil ||
			(!strings.Contains(err.Error(), "timeout") &&
				!strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
	}
	if err == nil {
		client.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

// finishRequest 处理响应，将数据写入请求结构体
func (client *Client) finishRequest(reply resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logger.Error(err)
		}
	}()
	request := <-client.waitingReqs
	if request == nil {
		return
	}
	request.reply = reply
	if request.waiting != nil {
		request.waiting.Done()
	}
}

// handleRead 读取服务端的响应
func (client *Client) handleRead() {
	ch := parser.ParseStream(client.conn)
	for payload := range ch {
		if payload.Err != nil {
			status := atomic.LoadInt32(&client.status)
			if status == closed {
				return
			}
			client.reconnect()
			return
		}
		client.finishRequest(payload.Data)
	}
}
