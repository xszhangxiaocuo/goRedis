package resp

type Reply interface {
	ToBytes() []byte //通信使用字节流
}
