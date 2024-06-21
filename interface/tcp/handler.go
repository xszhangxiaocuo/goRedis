package tcp

import (
	"context"
	"github.com/panjf2000/gnet/v2"
)

// Handler tcp连接处理器
type Handler interface {
	Handler(ctx context.Context, conn gnet.Conn)
	Close() error
}
