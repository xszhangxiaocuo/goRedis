// Package reply 错误响应
package reply

// UnknownErrReply 未知错误响应
type UnknownErrReply struct {
}

var unknownErrReply = new(UnknownErrReply)

var unknownErrBytes = []byte("-ERR unknown\r\n")

func NewUnknownErrReply() *UnknownErrReply {
	return unknownErrReply
}

func (u *UnknownErrReply) Error() string {
	return "ERR unknown"
}

func (u *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

// ArgNumErrReply 数组错误响应
type ArgNumErrReply struct {
	Cmd string
}

func NewArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		cmd,
	}
}

func (a *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + a.Cmd + "' command"
}

func (a *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + a.Cmd + "' command\r\n")
}

// SyntaxErrReply 语法错误响应
type SyntaxErrReply struct {
}

var syntaxErrReply = new(SyntaxErrReply)
var syntaxErrBytes = []byte("-ERR syntax error\r\n")

func NewSyntaxErrReply() *SyntaxErrReply {
	return syntaxErrReply
}

func (s *SyntaxErrReply) Error() string {
	return "ERR syntax error"
}

func (s *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

// WrongTypeErrReply 数据类型错误响应
type WrongTypeErrReply struct {
}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

func (w *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

func (w *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

// ProtocolErrReply 协议错误响应
type ProtocolErrReply struct {
	Msg string
}

func NewProtocolErrReply(msg string) *ProtocolErrReply {
	return &ProtocolErrReply{
		msg,
	}
}

func (p ProtocolErrReply) Error() string {
	return "ERR Protocol error: '" + p.Msg + "'"
}

func (p ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + p.Msg + "'\r\n")
}
