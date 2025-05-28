package database

import (
	"goRedis/interface/database"
	interDict "goRedis/interface/meta/dict"
	"goRedis/interface/resp"
	"goRedis/lib/logger"
	"goRedis/lib/utils"
	"goRedis/meta/dict"
	"goRedis/resp/reply"
	"strings"
)

// RedisDb 缓存数据库内核
type RedisDb struct {
	id     int                    // 数据库编号
	data   interDict.Dict         // 数据库存储的键值对
	addAof func(database.CmdLine) // 用于添加AOF命令行的函数
}

func NewRedisDb() *RedisDb {
	return &RedisDb{
		data:   dict.NewSyncDict(),
		addAof: func(line database.CmdLine) {},
	}
}

func (db *RedisDb) Exec(conn resp.Connection, cmdLine database.CmdLine) resp.Reply {
	cmdName := strings.ToLower(string(cmdLine[0])) // 第一行是命令名，后面的都是参数
	cmdName = strings.ToLower(cmdName)
	cmd, ok := cmdTable[cmdName]
	if !ok { // 未找到命令
		return reply.NewStandardErrReply("ERR unknown command '" + cmdName + "'")
	}
	if !utils.ValidateArgs(cmdLine, cmd.args) { // 参数个数不匹配
		return reply.NewArgNumErrReply(cmdName)
	}

	return cmd.execFunc(conn, db, cmdLine[1:])
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

func (db *RedisDb) AddAof(line database.CmdLine) {
	if db.addAof != nil {
		db.addAof(line)
	}
}

func (db *RedisDb) SetAddAof(fn func(database.CmdLine)) {
	db.addAof = fn
}
