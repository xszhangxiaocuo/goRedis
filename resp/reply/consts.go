// Package reply 固定写死的响应
package reply

// PongReply pong响应
type PongReply struct {
}

var pongReply = new(PongReply)

var pongbytes = []byte("+PONG\r\n")

func NewPongReply() *PongReply {
	return pongReply
}

func (p *PongReply) ToBytes() []byte {
	return pongbytes
}

// OkReply ok响应
type OkReply struct {
}

var okReply = new(OkReply)

var okbytes = []byte("+OK\r\n")

func NewOkReply() *OkReply {
	return okReply
}

func (o *OkReply) ToBytes() []byte {
	return okbytes
}

// NullBulkReply NULL响应
type NullBulkReply struct {
}

var nullBulkReply = new(NullBulkReply)

var nullbytes = []byte("$-1\r\n") //表示NULL，而不是空字符串

func NewNullBulkReply() *NullBulkReply {
	return nullBulkReply
}

func (n *NullBulkReply) ToBytes() []byte {
	return nullbytes
}

// EmptyBulkReply 空字符串响应
type EmptyBulkReply struct {
}

var emptyBulkReply = new(EmptyBulkReply)

var emptyBulkbytes = []byte("$0\r\n\r\n") //表示空字符串

func NewEmptyBulkReply() *EmptyBulkReply {
	return emptyBulkReply
}

func (n *EmptyBulkReply) ToBytes() []byte {
	return emptyBulkbytes
}

// EmptyMultiBulkReply 空数组响应
type EmptyMultiBulkReply struct {
}

var emptyMultiBulkReply = new(EmptyMultiBulkReply)

var emptyMultiBulkbytes = []byte("*-1\r\n") //表示空数组

func NewEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return emptyMultiBulkReply
}

func (n *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkbytes
}

// NoReply 空响应
type NoReply struct {
}

var noReply = new(NoReply)

var nobytes = []byte("") //表示响应为空（不是代表空字符串，而是响应内容为空）

func NewNoReply() *NoReply {
	return noReply
}

func (n *NoReply) ToBytes() []byte {
	return nobytes
}
