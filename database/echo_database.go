package database

import (
	"goRedis/interface/resp"
	"goRedis/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e *EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.NewMultiBulkReply(args)
}

func (e *EchoDatabase) Close() error {
	return nil
}

func (e *EchoDatabase) AfterClientClose(client resp.Connection) error {
	return nil
}
