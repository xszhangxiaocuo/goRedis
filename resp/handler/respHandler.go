package handler

import (
	"context"
	"errors"
	"goRedis/database"
	dbinterface "goRedis/interface/database"
	"goRedis/lib/logger"
	"goRedis/resp/connection"
	"goRedis/resp/parser"
	"goRedis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

const ErrClosed = "use of closed network connection" // 使用了一个已经关闭的连接

type RESPHandler struct {
	activeConn sync.Map             // 当前存活的连接
	db         dbinterface.Database // database接口类，redis业务层
	closing    atomic.Bool          // 用于判断当前是否处于关闭状态
}

func NewRESPHandler() *RESPHandler {
	// TODO: 实现database接口类
	db := database.NewEchoDatabase()
	return &RESPHandler{
		db: db,
	}
}

// Handler 处理客户端连接
func (r *RESPHandler) Handler(ctx context.Context, conn net.Conn) {
	if r.closing.Load() { // 如果当前处于关闭状态，关闭连接
		_ = conn.Close()
	}
	client := connection.NewRESPConn(conn) // 新建一个客户端连接
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(conn) // 解析客户端请求
	for payload := range ch {      // 不断读取管道中的数据，即客户端请求
		if payload.Err != nil { // 解析出错
			// 如果是EOF或者连接被关闭，关闭连接
			if errors.Is(payload.Err, io.EOF) ||
				errors.Is(payload.Err, io.ErrUnexpectedEOF) ||
				strings.Contains(payload.Err.Error(), ErrClosed) {
				r.closeClient(client) // 关闭客户端连接
				logger.Info("Connection closed: " + client.RemoteAddr().String())
				return
			}
			// 协议错误
			errReply := reply.NewProtocolErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("Connection closed: " + client.RemoteAddr().String())
				return
			}
			// 只是协议格式造成的解析错误，继续处理下一个请求
			continue
		}
		// 处理正常请求
		if payload.Data == nil { // 指令为空
			logger.Info("request is nil: " + client.RemoteAddr().String())
			continue
		}
		//客户端发送的指令必须是二维数组格式的
		mbreply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok { // 类型错误
			logger.Error("need multi bulk reply")
			continue
		}
		result := r.db.Exec(client, mbreply.Args) // 执行指令
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(reply.NewUnknownErrReply().ToBytes())
		}
	}
}

// Close 关闭协议层
func (r *RESPHandler) Close() error {
	logger.Info("handler shutting down")
	r.closing.Store(true)                                  // 设置关闭状态
	r.activeConn.Range(func(key, value interface{}) bool { // 关闭所有连接
		client := value.(*connection.RESPConn)
		_ = client.Close()
		return true
	})
	r.db.Close() // 关闭redis业务层
	return nil
}

// closeClient 关闭一个客户端连接
func (r *RESPHandler) closeClient(client *connection.RESPConn) {
	_ = client.Close()            // 关闭客户端连接
	r.db.AfterClientClose(client) // 关闭后连接后的清理操作
	r.activeConn.Delete(client)   // 删除连接
}
