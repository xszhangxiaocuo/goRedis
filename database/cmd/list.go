package cmd

import (
	"goRedis/database"
	interdb "goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/meta/list"
	"goRedis/resp/reply"
	"strconv"
)

func init() {
	database.RegisterCommand("lpush", LPush, -3)
	database.RegisterCommand("rpush", RPush, -3)
	database.RegisterCommand("lrange", LRange, 4)
	database.RegisterCommand("lpop", LPop, -2)
	database.RegisterCommand("rpop", RPop, -2)
	database.RegisterCommand("llen", LLen, 2)
}

// LPush 将所有指定的值插入存储在key的列表的头部。如果key不存在，则在执行推送操作之前将其创建为空列表。
func LPush(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		data := list.NewQuickList()
		for i := 1; i < len(args); i++ {
			data.Insert(0, args[i])
		}
		entity = interdb.NewDataEntity(data)
		db.PutEntity(key, entity)
		return reply.NewIntReply(int64(data.Len()))
	}

	data, ok := entity.Data.(*list.QuickList)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	for i := 1; i < len(args); i++ {
		data.Insert(0, args[i])
	}
	db.AddAof(utils.ToCmdLine3("lpush", args...))
	return reply.NewIntReply(int64(data.Len()))
}

// RPush 将所有指定的值插入存储在key的列表的尾部。如果key不存在，则在执行推送操作之前将其创建为空列表。
func RPush(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		data := list.NewQuickList()
		for i := 1; i < len(args); i++ {
			data.Add(args[i])
		}
		entity = interdb.NewDataEntity(data)
		db.PutEntity(key, entity)
		return reply.NewIntReply(int64(data.Len()))
	}

	data, ok := entity.Data.(*list.QuickList)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	for i := 1; i < len(args); i++ {
		data.Add(args[i])
	}
	db.AddAof(utils.ToCmdLine3("rpush", args...))
	return reply.NewIntReply(int64(data.Len()))
}

// LPop 移除并返回存储在key的列表的第一个元素,当提供可选的count参数时，回复将由最多count个元素组成，这取决于列表的长度。
func LPop(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewEmptyMultiBulkReply()
	}

	data, ok := entity.Data.(*list.QuickList)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	count := 1
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(string(args[1]))
		if err != nil {
			return reply.NewStandardErrReply("ERR value is not an integer or out of range")
		}
	}
	if count < 0 {
		return reply.NewStandardErrReply("ERR value is not an integer or out of range")
	}
	result := make([][]byte, 0)
	for i := 0; i < count; i++ {
		if data.Len() == 0 {
			break
		}
		result = append(result, data.Remove(0).([]byte))
	}
	if len(result) == 0 {
		return reply.NewEmptyMultiBulkReply()
	}
	db.AddAof(utils.ToCmdLine3("lpop", args...))
	return reply.NewMultiBulkReply(result)
}

// RPop 移除并返回存储在key的列表的最后一个元素,当提供可选的count参数时，回复将由最多count个元素组成，这取决于列表的长度。
func RPop(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewNullBulkReply()
	}

	data, ok := entity.Data.(*list.QuickList)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	count := 1
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(string(args[1]))
		if err != nil {
			return reply.NewStandardErrReply("ERR value is not an integer or out of range")
		}
	}
	if count < 0 {
		return reply.NewStandardErrReply("ERR value is not an integer or out of range")
	}
	result := make([][]byte, 0)
	for i := 0; i < count; i++ {
		if data.Len() == 0 {
			break
		}
		result = append(result, data.Remove(data.Len()-1).([]byte))
	}
	if len(result) == 0 {
		return reply.NewNullBulkReply()
	}
	db.AddAof(utils.ToCmdLine3("rpop", args...))
	return reply.NewMultiBulkReply(result)
}

// LLen 返回存储在key的列表的长度
func LLen(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewIntReply(0)
	}

	data, ok := entity.Data.(*list.QuickList)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	return reply.NewIntReply(int64(data.Len()))

}

// LRange 遍历范围为[start,stop]
func LRange(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.NewNoReply()
	}
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return reply.NewStandardErrReply("ERR type error")
	}
	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return reply.NewStandardErrReply("ERR type error")
	}

	data, ok := entity.Data.(*list.QuickList)
	if !ok {
		return reply.NewStandardErrReply("ERR type error")
	}
	size := data.Len()
	if start >= size {
		return reply.NewNullBulkReply()
	}
	start, stop = changeRange(start, stop, size)
	items := data.Range(start, stop)
	result := make([][]byte, len(items))
	for i, item := range items {
		result[i] = item.([]byte)
	}
	return reply.NewMultiBulkReply(result)
}

// changeRange 将传入的下标转换为合法范围
func changeRange(start int, stop int, size int) (int, int) {
	if start < -1*size {
		start = 0
	} else if start < 0 {
		start = size + start
	}

	if stop < -1*size {
		stop = 0
	} else if stop < 0 {
		stop = size + stop + 1
	} else if stop < size {
		stop = stop + 1
	} else {
		stop = size
	}
	if stop < start {
		stop = start
	}
	return start, stop
}
