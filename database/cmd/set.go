package cmd

import (
	"goRedis/database"
	database2 "goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/meta/set"
	"goRedis/resp/reply"
)

func init() {
	database.RegisterCommand("sadd", SAdd, -3)
	database.RegisterCommand("srem", SRem, -3)
	database.RegisterCommand("sismember", SIsMember, 3)
}

// SAdd 向集合添加一个或多个成员
func SAdd(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		entity = database2.NewDataEntity(set.NewSet())
		db.PutEntity(key, entity)
	}
	result := 0
	for i, arg := range args {
		if i == 0 {
			continue
		}
		data, ok := entity.Data.(*set.Set)
		if !ok {
			return reply.NewStandardErrReply("ERR type error")
		}
		result += data.Add(string(arg))
	}
	db.AddAof(utils.ToCmdLine3("sadd", args...))
	return reply.NewIntReply(int64(result))
}

// SRem 移除集合中一个或多个成员
func SRem(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewIntReply(0)
	}
	result := 0
	for i, arg := range args {
		if i == 0 {
			continue
		}
		data, ok := entity.Data.(*set.Set)
		if !ok {
			return reply.NewStandardErrReply("ERR type error")
		}
		result += data.Remove(string(arg))
	}
	db.AddAof(utils.ToCmdLine3("srem", args...))
	return reply.NewIntReply(int64(result))
}

// SIsMember 判断成员元素是否是集合的成员
func SIsMember(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewIntReply(0)
	}
	data, ok := entity.Data.(*set.Set)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	if data.Has(string(args[1])) {
		return reply.NewIntReply(1)
	}
	return reply.NewIntReply(0)
}
