package cmd

import (
	"goRedis/database"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/resp/reply"
	"strings"
)

func init() {
	database.RegisterCommand("client", Client, -2)
	registerClientCmd("setname", 2)
	registerClientCmd("getname", 1)
}

var clientCmdTable map[string]int = make(map[string]int)

// Client 客户端相关命令
// 包含client setname、client getname等
func Client(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	cmdName := string(args[0])
	cmdName = strings.ToLower(cmdName)
	if !utils.ValidateArgs(args, clientCmdTable[cmdName]) { // 参数个数不匹配
		return reply.NewArgNumErrReply(cmdName)
	}
	switch cmdName {
	case "setname":
		return ClientSetName(client, db, args[1:])
	case "getname":
		return ClientGetName(client, db, args[1:])
	default:
		return reply.NewStandardErrReply("ERR unknown command 'client " + cmdName + "'")
	}
}

// ClientSetName 修改当前连接的名称
func ClientSetName(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	name := args[0]
	client.SetName(name)
	return reply.NewOkReply()
}

// ClientGetName 修改当前连接的名称
func ClientGetName(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	name := client.GetName()
	if len(name) == 0 {
		return reply.NewNullBulkReply()
	}
	return reply.NewBulkReply(name)
}

func registerClientCmd(cmdName string, args int) {
	clientCmdTable[cmdName] = args
}
