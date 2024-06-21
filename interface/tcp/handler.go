package tcp

import (
	"context"
	"goRedis/interface/resp"
)

// Handler tcp连接处理器
type Handler interface {
	Handler(ctx context.Context, conn resp.Connection)
	Close() error
}
