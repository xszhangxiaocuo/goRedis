package cmd

import (
	"goRedis/database"
	"goRedis/interface/resp"
	"goRedis/resp/reply"
)

func init() {
	database.RegisterCommand("ping", Ping, -1)
	database.RegisterCommand("echo", Echo, 2)
}

// Ping 传入的args不包括命令名，只传入参数
func Ping(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	if len(args) > 0 {
		return reply.NewStatusReply(string(args[0]))
	}
	return reply.NewPongReply()
}

// Echo 返回传入的参数
func Echo(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	return reply.NewStatusReply(string(args[0]))
}
