package cmd

import (
	"goRedis/database"
	idatabase "goRedis/interface/database"
	idict "goRedis/interface/meta/dict"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/meta/dict"
	"goRedis/resp/reply"
)

func init() {
	database.RegisterCommand("hset", HSet, 4)
	database.RegisterCommand("hget", HGet, 3)
	database.RegisterCommand("hdel", HDel, -3)
}

// HSet 向哈希表中添加一个字段
func HSet(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		entity = idatabase.NewDataEntity(dict.NewSyncDict())
		db.PutEntity(key, entity)
	}
	data, ok := entity.Data.(idict.Dict)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	field := string(args[1])
	db.AddAof(utils.ToCmdLine3("hset", args...))
	return reply.NewIntReply(int64(data.Put(field, args[2])))
}

// HGet 获取存储在哈希表中指定字段的值
func HGet(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewNullBulkReply()
	}
	data, ok := entity.Data.(idict.Dict)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	field := string(args[1])
	val, exists := data.Get(field)
	if !exists {
		return reply.NewNullBulkReply()
	}
	return reply.NewBulkReply(val.([]byte))
}

// HDel 删除存储在哈希表中指定字段的值
func HDel(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewIntReply(0)
	}
	data, ok := entity.Data.(idict.Dict)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	result := 0
	for i, arg := range args {
		if i == 0 {
			continue
		}
		result += data.Remove(string(arg))
	}
	if result > 0 {
		db.AddAof(utils.ToCmdLine3("hdel", args...))
	}
	return reply.NewIntReply(int64(result))
}
