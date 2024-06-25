package cmd

import (
	"goRedis/database"
	"goRedis/interface/resp"
	"goRedis/lib/wildcard"
	"goRedis/resp/reply"
)

func init() {
	database.RegisterCommand("del", Del, -2)
	database.RegisterCommand("exists", Exists, -2)
	database.RegisterCommand("flushdb", FlushDb, -1)
	database.RegisterCommand("type", Type, 2)
	database.RegisterCommand("rename", Rename, 3)
	database.RegisterCommand("renamenx", RenameNX, 3)
	database.RegisterCommand("keys", Keys, 2)
}

// Del 删除多个键值对，返回成功删除的个数
func Del(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}
	count := db.RemoveAll(keys...)
	return reply.NewIntReply(int64(count))
}

func Exists(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	count := 0
	for _, arg := range args {
		_, existed := db.GetEntity(string(arg))
		if existed {
			count++
		}
	}
	return reply.NewIntReply(int64(count))
}

// FlushDb 清空数据库 TODO: 参数：SYNC同步刷新数据库，ASYNC异步刷新数据库
func FlushDb(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	db.Close()
	return reply.NewOkReply()
}

// Type 返回存储在key的值的类型的字符串表示，可以返回的不同类型有：string、list、set、zset、hash
func Type(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	entity, existed := db.GetEntity(string(args[0]))
	if !existed {
		return reply.NewStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.NewStatusReply("string")
		// TODO: 其他类型
	default:
		return reply.NewUnknownErrReply()
	}
}

// Rename 重命名一个key，如果新key已经存在，则进行覆盖
func Rename(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.NewStandardErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	return reply.NewOkReply()
}

// RenameNX 重命名一个key，如果新key已经存在，则什么都不做
func RenameNX(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	if _, ok := db.GetEntity(dest); ok { // 如果dest已经存在，不做任何操作
		return reply.NewIntReply(0)
	}

	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.NewStandardErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	return reply.NewIntReply(1)
}

// Keys 查找所有符合给定模式 pattern 的 key
func Keys(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0])) // 解析通配符
	result := make([][]byte, 0)
	db.GetData().ForEach(func(key string, value interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.NewMultiBulkReply(result)
}
