package database

import (
	"fmt"
	"goRedis/config"
	"goRedis/interface/resp"
	"goRedis/lib/logger"
	"goRedis/resp/reply"
	"strconv"
	"strings"
)

type Database struct {
	dbSet []*RedisDb
}

func NewDataBase() *Database {
	database := &Database{}
	if config.Properties.Databases <= 0 { // 默认16个数据库
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*RedisDb, config.Properties.Databases)
	for i := 0; i < config.Properties.Databases; i++ {
		db := NewRedisDb()
		db.SetId(i)
		database.dbSet[i] = db
	}
	return database
}

func (db *Database) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(fmt.Sprintf("panic: %v", err))
		}
	}()

	cmdName := string(args[0])
	cmdName = strings.ToLower(cmdName)
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.NewArgNumErrReply("select")
		}
		return Select(client, db, args[1:])
	}
	dbIndex := client.GetDBIndex()
	if dbIndex < 0 || dbIndex >= len(db.dbSet) {
		return reply.NewStandardErrReply("ERR DB index out of range")
	}
	return db.dbSet[dbIndex].Exec(client, args)
}

func (db *Database) Close() error {
	return nil
}

func (db *Database) AfterClientClose(client resp.Connection) error {
	return nil
}

// Select 选择数据库
func Select(conn resp.Connection, db *Database, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.NewStandardErrReply("ERR invalid DB index")
	}
	if dbIndex < 0 || dbIndex >= len(db.dbSet) {
		return reply.NewStandardErrReply("ERR DB index out of range")
	}
	conn.SelectDB(dbIndex)
	return reply.NewOkReply()
}
