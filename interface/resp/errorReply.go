package resp

// ErrorReply 错误回复接口
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
