package atomic

import "sync/atomic"

//atomic包中没有bool类型，为了方便使用，将uint32包装一层作为bool使用，0表示false，非0表示true

// Boolean 是一个布尔值，它的所有操作都是原子的
type Boolean uint32

// Get 以原子方式读取值
func (b *Boolean) Get() bool {
	// atomic.LoadUint32 以原子方式加载 *uint32 类型的值
	// 如果加载的值不等于 0，则返回 true，否则返回 false
	return atomic.LoadUint32((*uint32)(b)) != 0
}

// Set 以原子方式写入值
func (b *Boolean) Set(v bool) {
	if v {
		// 如果 v 为 true，则以原子方式将 *uint32 类型的值设置为 1
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		// 如果 v 为 false，则以原子方式将 *uint32 类型的值设置为 0
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
