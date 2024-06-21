package dict

import (
	"goRedis/interface/meta/dict"
)

// Dict 普通字典
type Dict struct {
	m map[string]any
}

func NewDict() *Dict {
	return &Dict{
		m: make(map[string]any),
	}
}

func (dict *Dict) Get(key string) (val any, exists bool) {
	val, exists = dict.m[key]
	return
}

func (dict *Dict) Len() int {
	return len(dict.m)
}

func (dict *Dict) Put(key string, val any) (result int) {
	_, existed := dict.m[key]
	dict.m[key] = val
	result = 1   // 插入操作
	if existed { // 更新操作
		result = 0
	}
	return
}

func (dict *Dict) PutIfAbsent(key string, val any) (result int) {
	_, existed := dict.m[key]
	if existed { // key存在，不插入
		return 0
	}
	dict.m[key] = val
	return 1
}

func (dict *Dict) PutIfExist(key string, val any) (result int) {
	_, existed := dict.m[key]
	if existed { // key存在，更新
		dict.m[key] = val
		return 1
	}
	return 0
}

func (dict *Dict) Remove(key string) (result int) {
	_, exists := dict.m[key]
	if exists {
		delete(dict.m, key)
		return 1
	}
	return 0
}

func (dict *Dict) Keys() []string {
	result := make([]string, 0)
	for k := range dict.m {
		result = append(result, k)
	}
	return result
}

func (dict *Dict) ForEach(consumer dict.Consumer) {
	for k, v := range dict.m {
		if !consumer(k, v) {
			break
		}
	}
}

func (dict *Dict) RandomKeys(num int) []string {
	result := make([]string, 0)
	i := 0
	for k := range dict.m {
		result = append(result, k)
		i++
		if i == num {
			break
		}
	}
	return result
}

func (dict *Dict) RandomDistinctKeys(num int) []string {
	result := make([]string, 0)
	i := 0
	for k := range dict.m {
		// 如果result中已经存在k，则跳过
		for _, v := range result {
			if v == k {
				result = append(result, k)
				i++
				break
			}
		}
		if i == num {
			break
		}
	}
	return result
}

func (dict *Dict) Clear() {
	*dict = *NewDict() // 将dict指针指向一个新的SyncDict对象，指针指向的旧对象会被GC回收
}
