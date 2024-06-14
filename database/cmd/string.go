package cmd

import (
	"goRedis/database"
	interdb "goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/resp/reply"
)

func init() {
	database.RegisterCommand("get", Get, 2)
	database.RegisterCommand("set", Set, -3)
	database.RegisterCommand("setnx", SetNX, 3)
	database.RegisterCommand("getset", GetSet, 3)
	database.RegisterCommand("strlen", StrLen, 2)
}

// Get 获取key的值，如果key不存在则返回nil，如果key的值不是字符串则返回错误，字符串以[]byte形式存储
func Get(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewNullBulkReply()
	}
	value, ok := entity.Data.([]byte)
	if !ok { // 类型错误
		return reply.NewStandardErrReply("type error")
	}
	return reply.NewBulkReply(value)
}

// Set 设置key的值为value TODO: 暂时只支持'set key value'，其他可选参数后续支持
func Set(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	db.PutEntity(key, interdb.NewDataEntity(value))
	return reply.NewOkReply()
}

// SetNX 设置key的值为value，如果key已经存在则不做任何操作
func SetNX(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	result := db.PutIfAbsent(key, interdb.NewDataEntity(value))
	return reply.NewIntReply(int64(result))
}

// GetSet 设置key的值为value，并返回key的旧值
func GetSet(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity, exists := db.GetEntity(key)
	db.PutEntity(key, interdb.NewDataEntity(value))
	if !exists {
		return reply.NewNullBulkReply()
	}
	v, ok := entity.Data.([]byte)
	if !ok { // 类型错误
		return reply.NewStandardErrReply("type error")
	}
	return reply.NewBulkReply(v)
}

// StrLen 返回key的字符串值的长度
func StrLen(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists { // key不存在，返回0
		return reply.NewIntReply(0)
	}
	value, ok := entity.Data.([]byte)
	if !ok { // 类型错误
		return reply.NewStandardErrReply("ERR type error")
	}
	return reply.NewIntReply(int64(len(value)))
}
