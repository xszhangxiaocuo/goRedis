package list

// Expected 检查传入的参数是否符合预期
type Expected func(a any) bool

// Consumer 遍历列表时的回调函数
type Consumer func(i int, v any) bool

type List interface {
	Add(val any)
	Get(index int) (val any)
	Set(index int, val any)
	Insert(index int, val any)
	Remove(index int) (val any)
	RemoveLast() (val any)
	RemoveAllByVal(expected Expected) int
	RemoveByVal(expected Expected, count int) int
	ReverseRemoveByVal(expected Expected, count int) int
	Len() int
	ForEach(consumer Consumer)
	Contains(expected Expected) bool
	Range(start int, stop int) []any
}
