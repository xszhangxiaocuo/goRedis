package resp

import "github.com/panjf2000/gnet/v2"

type Connection interface {
	Write([]byte) error
	Close() error
	GetDBIndex() int
	SelectDB(int)
	SetName(name []byte)
	GetName() []byte
	GetConn() gnet.Conn
}
