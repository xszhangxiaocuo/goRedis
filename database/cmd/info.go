package cmd

import (
	"goRedis/database"
	"goRedis/interface/resp"
	"goRedis/resp/reply"
)

func init() {
	database.RegisterCommand("info", Info, -1)
}

// Info 返回服务器信息 TODO: 参数：keyspace、clients、memory、persistence、stats、replication、cpu、commandstats、cluster、keyspace等
func Info(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	info := make([]byte, 0)
	for _, arg := range args {
		str := string(arg)
		switch str {
		case "keyspace":
			info = append(info, []byte("# Keyspace\r\ndb0:keys=1,expires=0,avg_ttl=0\r\n")...) // TODO: 暂时写死
		}
	}
	return reply.NewBulkReply(info)
}
