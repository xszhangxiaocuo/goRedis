package database

import (
	"goRedis/interface/resp"
	"strings"
)

type ExecFunc func(db *RedisDb, args [][]byte) resp.Reply // 命令执行函数，接收数据库和参数，返回RESP协议回复
var cmdTable = make(map[string]*command)                  // 命令名 -> command

type command struct {
	name     string   // 命令名
	execFunc ExecFunc // 命令执行函数
	args     int      // 参数个数
}

func RegisterCommand(name string, execFunc ExecFunc, args int) {
	name = strings.ToLower(name) // 命令名不区分大小写
	cmdTable[name] = &command{
		name:     name,
		execFunc: execFunc,
		args:     args,
	}
}
