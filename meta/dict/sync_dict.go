package dict

import (
	"goRedis/interface/meta/dict"
	"sync"
)

// SyncDict Redis核心数据结构之一，线程安全的字典，最底层的用于存储键值对的数据结构
type SyncDict struct {
	m sync.Map
}

func NewSyncDict() *SyncDict {
	return &SyncDict{}
}

func (dict *SyncDict) Get(key string) (val any, exists bool) {
	val, exists = dict.m.Load(key)
	return
}

func (dict *SyncDict) Len() int {
	length := 0
	dict.m.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

func (dict *SyncDict) Put(key string, val any) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	result = 1   // 插入操作
	if existed { // 更新操作
		result = 0
	}
	return
}

func (dict *SyncDict) PutIfAbsent(key string, val any) (result int) {
	_, existed := dict.m.Load(key)
	if existed { // key存在，不插入
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

func (dict *SyncDict) PutIfExist(key string, val any) (result int) {
	_, existed := dict.m.Load(key)
	if existed { // key存在，更新
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

func (dict *SyncDict) Remove(key string) (result int) {
	_, exists := dict.m.Load(key)
	if exists {
		dict.m.Delete(key)
		return 1
	}
	return 0
}

func (dict *SyncDict) Keys() []string {
	result := make([]string, 0)
	dict.m.Range(func(key, value any) bool {
		result = append(result, key.(string))
		return true
	})
	return result
}

func (dict *SyncDict) ForEach(consumer dict.Consumer) {
	dict.m.Range(func(key, value any) bool {
		consumer(key.(string), value)
		return true
	})
}

func (dict *SyncDict) RandomKeys(num int) []string {
	result := make([]string, 0)
	for i := 0; i < num; i++ {
		dict.m.Range(func(key, value any) bool {
			result = append(result, key.(string))
			return false
		})
	}
	return result
}

func (dict *SyncDict) RandomDistinctKeys(num int) []string {
	result := make([]string, 0)
	i := 0
	dict.m.Range(func(key, value any) bool {
		result = append(result, key.(string))
		i++
		if i == num {
			return false
		}
		return true
	})
	return result
}

func (dict *SyncDict) Clear() {
	*dict = *NewSyncDict() // 将dict指针指向一个新的SyncDict对象，指针指向的旧对象会被GC回收
}
