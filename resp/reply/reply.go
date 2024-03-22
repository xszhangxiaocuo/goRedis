// Package reply 自定义响应
package reply

import (
	"bytes"
	"goRedis/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1") //NULL
	CRLF               = "\r\n"        //换行
)

// BulkReply 单个字符串响应
type BulkReply struct {
	Arg []byte
}

func NewBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		arg,
	}
}

// ToBytes 拼接字符串类型数据的响应，示例：$5\r\nhello\r\n
func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 { //字符串为空返回NULL
		return NewNullBulkReply().ToBytes()
	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

// MultiBulkReply 数组响应
type MultiBulkReply struct {
	Args [][]byte
}

func NewMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		args,
	}
}

func (m *MultiBulkReply) ToBytes() []byte {
	var buf bytes.Buffer
	argLen := len(m.Args)
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range m.Args {
		if arg == nil { //写入一个NULL
			buf.WriteString(string(nullBulkReplyBytes) + CRLF)
		} else { //写入一个字符串
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

// StatusReply 状态响应
type StatusReply struct {
	Status string
}

func NewStatusReply(status string) *StatusReply {
	return &StatusReply{
		status,
	}
}

func (s *StatusReply) ToBytes() []byte {
	return []byte("+" + s.Status + CRLF)
}

// IntReply 数字响应
type IntReply struct {
	Code int64
}

func NewIntReply(code int64) *IntReply {
	return &IntReply{
		code,
	}
}

func (i *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + CRLF)
}

// StandardErrReply 标准错误响应
type StandardErrReply struct {
	Msg string
}

func NewStandardErrReply(msg string) *StandardErrReply {
	return &StandardErrReply{
		msg,
	}
}

func (s *StandardErrReply) Error() string {
	return s.Msg
}

func (s *StandardErrReply) ToBytes() []byte {
	return []byte("-" + s.Msg + CRLF)
}

// IsErrReply 判断传入的reply是否是ErrReply
func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
