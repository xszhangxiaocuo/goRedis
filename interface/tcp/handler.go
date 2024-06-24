package tcp

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"goRedis/interface/resp"
)

// Handler tcp连接处理器
type Handler interface {
	Handler(ctx context.Context, conn resp.Connection, data []byte) (result []byte, action gnet.Action)
	Close() error
}
