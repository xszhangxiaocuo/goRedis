package database

import "goRedis/interface/resp"

type CmdLine = [][]byte //传入的参数都是字节数组，故使用一个别名进行替换

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close() error
	AfterClientClose(client resp.Connection) error //在客户端关闭后可能需要进行一些清理操作
}

type DataEntity struct {
	Data any
}
