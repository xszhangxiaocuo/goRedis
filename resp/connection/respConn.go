package connection

import (
	"goRedis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

type RESPConn struct {
	conn         net.Conn
	waitingReply wait.Wait  //等待所有处理完成
	mu           sync.Mutex //每个连接要加锁
	selectedDB   int        //标记当前连接正在使用的数据库id
}

// NewRESPConn 创建一个新的RESPConn
func NewRESPConn(conn net.Conn) *RESPConn {
	return &RESPConn{
		conn: conn,
	}
}

// RemoteAddr 查看远程连接的addr
func (r *RESPConn) RemoteAddr() net.Addr {
	return r.conn.RemoteAddr()
}

// Close 关闭连接
func (r *RESPConn) Close() error {
	r.waitingReply.WaitWithTimeout(10 * time.Second) //等待10s超时
	return r.conn.Close()
}

// Write 向连接写入数据
func (r *RESPConn) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	r.mu.Lock()
	r.waitingReply.Add(1)
	defer func() {
		r.waitingReply.Done()
		r.mu.Unlock()
	}()
	_, err := r.conn.Write(bytes)

	return err
}

// GetDBIndex 获取当前连接正在使用的数据库id
func (r *RESPConn) GetDBIndex() int {
	return r.selectedDB
}

// SelectDB 选择数据库
func (r *RESPConn) SelectDB(id int) {
	r.selectedDB = id
}
