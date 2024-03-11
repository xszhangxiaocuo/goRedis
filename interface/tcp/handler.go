package tcp

import (
	"context"
	"net"
)

// Handler tcp连接处理器
type Handler interface {
	Handler(ctx context.Context, conn net.Conn)
	Close() error
}
