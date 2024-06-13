package database

import (
	"goRedis/interface/database"
	interDict "goRedis/interface/meta/dict"
	"goRedis/interface/resp"
	"goRedis/lib/logger"
	"goRedis/meta/dict"
	"goRedis/resp/reply"
	"strings"
)

// RedisDb Redis内核
type RedisDb struct {
	id   int            // 数据库编号
	data interDict.Dict // 数据库存储的键值对
}

func NewRedisDb() *RedisDb {
	return &RedisDb{
		data: dict.NewSyncDict(),
	}
}

func (db *RedisDb) Exec(conn resp.Connection, cmdLine database.CmdLine) resp.Reply {
	cmdName := strings.ToLower(string(cmdLine[0])) // 第一行是命令名，后面的都是参数
	cmdName = strings.ToLower(cmdName)
	cmd, ok := cmdTable[cmdName]
	if !ok { // 未找到命令
		return reply.NewStandardErrReply("ERR unknown command '" + cmdName + "'")
	}
	if !validateArgs(cmdLine, cmd.args) { // 参数个数不匹配
		return reply.NewArgNumErrReply(cmdName)
	}

	return cmd.execFunc(db, cmdLine[1:])
}

func (db *RedisDb) GetEntity(key string) (*database.DataEntity, bool) {
	val, exists := db.data.Get(key)
	if !exists {
		return nil, false
	}
	entity, ok := val.(*database.DataEntity)
	if !ok {
		logger.Error("value of key %s is not DataEntity", key)
		return nil, false
	}
	return entity, true
}

func (db *RedisDb) GetData() interDict.Dict {
	return db.data
}

func (db *RedisDb) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

func (db *RedisDb) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExist(key, entity)
}

func (db *RedisDb) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

func (db *RedisDb) Remove(key string) int {
	return db.data.Remove(key)
}

// RemoveAll 删除多个键值对，返回成功删除的数量
func (db *RedisDb) RemoveAll(keys ...string) int {
	count := 0
	for _, key := range keys {
		count += db.Remove(key)
	}
	return count
}

func (db *RedisDb) Close() error {
	db.data.Clear()
	return nil
}

func (db *RedisDb) AfterClientClose(client resp.Connection) error {
	return nil
}

func (db *RedisDb) SetId(id int) {
	db.id = id
}

// 校验参数个数，命令本身也算一个参数，所有参数个数至少为1
func validateArgs(args [][]byte, expected int) bool {
	argNum := len(args)
	if expected >= 0 { // 定长参数
		return argNum == expected
	}

	// 最少参数，比如-2表示至少2个参数
	return argNum >= -expected
}
