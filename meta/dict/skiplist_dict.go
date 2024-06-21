package dict

import (
	"github.com/zhangyunhao116/skipmap"
	"goRedis/interface/meta/dict"
)

// SkipListDict Redis核心数据结构之一，线程安全的字典，最底层的用于存储键值对的数据结构
type SkipListDict struct {
	m *skipmap.StringMap[any]
}

func NewSkipListDict() *SkipListDict {
	return &SkipListDict{
		m: skipmap.NewString[any](),
	}
}

func (dict *SkipListDict) Get(key string) (val any, exists bool) {
	val, exists = dict.m.Load(key)
	return
}

func (dict *SkipListDict) Len() int {
	length := 0
	dict.m.Range(func(key string, value any) bool {
		length++
		return true
	})
	return length
}

func (dict *SkipListDict) Put(key string, val any) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	result = 1   // 插入操作
	if existed { // 更新操作
		result = 0
	}
	return
}

func (dict *SkipListDict) PutIfAbsent(key string, val any) (result int) {
	_, existed := dict.m.Load(key)
	if existed { // key存在，不插入
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

func (dict *SkipListDict) PutIfExist(key string, val any) (result int) {
	_, existed := dict.m.Load(key)
	if existed { // key存在，更新
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

func (dict *SkipListDict) Remove(key string) (result int) {
	_, exists := dict.m.Load(key)
	if exists {
		dict.m.Delete(key)
		return 1
	}
	return 0
}

func (dict *SkipListDict) Keys() []string {
	result := make([]string, 0)
	dict.m.Range(func(key string, value any) bool {
		result = append(result, key)
		return true
	})
	return result
}

func (dict *SkipListDict) ForEach(consumer dict.Consumer) {
	dict.m.Range(func(key string, value any) bool {
		consumer(key, value)
		return true
	})
}

func (dict *SkipListDict) RandomKeys(num int) []string {
	result := make([]string, 0)
	for i := 0; i < num; i++ {
		dict.m.Range(func(key string, value any) bool {
			result = append(result, key)
			return false
		})
	}
	return result
}

func (dict *SkipListDict) RandomDistinctKeys(num int) []string {
	result := make([]string, 0)
	i := 0
	dict.m.Range(func(key string, value any) bool {
		result = append(result, key)
		i++
		if i == num {
			return false
		}
		return true
	})
	return result
}

func (dict *SkipListDict) Clear() {
	*dict = *NewSkipListDict() // 将dict指针指向一个新的SkipListDict对象，指针指向的旧对象会被GC回收
}
