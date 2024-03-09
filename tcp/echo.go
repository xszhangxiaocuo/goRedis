package tcp

import (
	"bufio"
	"context"
	"goRedis/lib/logger"
	"goRedis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// EchoClient 用于测试的客户端
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	if e.Waiting.WaitWithTimeout(10 * time.Second) { //超时等待10s
		logger.Warn("关闭连接超时")
	}
	_ = e.Conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Bool //用于判断当前是否处于关闭状态
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handler(ctx context.Context, conn net.Conn) {
	if handler.closing.Load() {
		_ = conn.Close()
	}
	//新建一个测试客户端
	client := &EchoClient{
		Conn: conn,
	}
	handler.activeConn.Store(client, struct{}{}) //只需要保存key，不需要value，用空结构体能节省空间
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF { //客户端连接已经关闭
				logger.Info("Connection close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
		}
		//增加一个事件计数
		client.Waiting.Add(1)
		b := []byte(msg)
		_, err = conn.Write(b)
		if err != nil {
			logger.Warn(err)
		}
		//事件完成，计数器减1
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("echoHandler shutting down...")
	handler.closing.Store(true)                          //设置当前handler正在进行关闭操作
	handler.activeConn.Range(func(key, value any) bool { //遍历map关闭所有客户端连接
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		return true
	})
	return nil
}
