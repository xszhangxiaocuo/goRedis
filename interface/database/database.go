package database

import "goRedis/interface/resp"

type CmdLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close() error
	AfterClientClose(client resp.Connection) error //在客户端关闭后可能需要进行一些操作
}

type DataEntity struct {
	Data any
}
