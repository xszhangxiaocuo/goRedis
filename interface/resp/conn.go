package resp

type Connection interface {
	Write([]byte) error
	GetDBIndex() int
	SelectDB(int)
	SetName(name []byte)
	GetName() []byte
}
