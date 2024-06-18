package cmd

import (
	"goRedis/database"
	interdb "goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/meta/list"
	"goRedis/resp/reply"
	"strconv"
)

func init() {
	database.RegisterCommand("lpush", LPush, -3)
	database.RegisterCommand("rpush", RPush, -3)
	database.RegisterCommand("lrange", LRange, 4)
}

// LPush 将所有指定的值插入存储在key的列表的头部。如果key不存在，则在执行推送操作之前将其创建为空列表。
func LPush(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		data := list.NewLinkedList()
		for i := 1; i < len(args); i++ {
			data.Insert(0, args[i])
		}
		entity = interdb.NewDataEntity(data)
		db.PutEntity(key, entity)
		return reply.NewIntReply(int64(data.Len()))
	}

	data, ok := entity.Data.(*list.LinkedList)
	if !ok {
		return reply.NewStandardErrReply("type error")
	}
	for i := 1; i < len(args); i++ {
		data.Insert(0, args[i])
	}
	return reply.NewIntReply(int64(data.Len()))
}

// RPush 将所有指定的值插入存储在key的列表的尾部。如果key不存在，则在执行推送操作之前将其创建为空列表。
func RPush(client resp.Connection, db *database.RedisDb, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		data := list.NewLinkedList()
		for i := 1; i < len(args); i++ {
			data.Add(args[i])
		}
		entity = interdb.NewDataEntity(data)
		db.PutEntity(key, entity)
		return reply.NewIntReply(int64(data.Len()))
	}

	data, ok := entity.Data.(*list.LinkedList)
	if !ok {
		return reply.NewStandardErrReply("type error")
	}
	for i := 1; i < len(args); i++ {
		data.Add(args[i])
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

	data, ok := entity.Data.(*list.LinkedList)
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
	return reply.NewListReply(result)
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
