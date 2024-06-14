package cmd

import (
	"goRedis/config"
	"goRedis/database"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/resp/reply"
	"strings"
)

func init() {
	database.RegisterCommand("config", Config, -2)
	registerConfigCmd("get", -2)
}

var configCmdTable map[string]int = make(map[string]int)

// Config 客户端相关命令
// 包含config get 等
func Config(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	cmdName := string(args[0])
	cmdName = strings.ToLower(cmdName)
	if !utils.ValidateArgs(args, configCmdTable[cmdName]) { // 参数个数不匹配
		return reply.NewArgNumErrReply(cmdName)
	}
	switch cmdName {
	case "get":
		return ConfigGet(client, db, args[1:])
	default:
		return reply.NewStandardErrReply("ERR unknown command 'config " + cmdName + "'")
	}
}

// ConfigGet 获取服务器配置信息
func ConfigGet(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	result := make([][]byte, 0)
	for _, arg := range args {
		cmdName := string(arg)
		cmdName = strings.ToLower(cmdName)
		result = append(result, []byte(cmdName))
		result = append(result, []byte(config.Properties.GetConfig(cmdName)))
	}

	return reply.NewMultiBulkReply(result)
}

func registerConfigCmd(cmdName string, args int) {
	configCmdTable[cmdName] = args
}
