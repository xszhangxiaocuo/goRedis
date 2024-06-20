package set

import (
	"goRedis/interface/meta/dict"
	sdict "goRedis/meta/dict"
)

// Set 基于字典实现的集合，线程安全。set本质上是一个字典，只不过字典的value是一个空结构体
type Set struct {
	dict dict.Dict
}

func NewSet(members ...string) *Set {
	set := &Set{
		dict: sdict.NewSyncDict(),
	}
	for _, member := range members {
		set.Add(member)
	}
	return set
}

func (set *Set) Add(val string) int {
	return set.dict.Put(val, nil)
}

func (set *Set) Remove(val string) int {
	result := set.dict.Remove(val)
	return result
}

// Has 检查val是否在set中
func (set *Set) Has(val string) bool {
	if set == nil || set.dict == nil {
		return false
	}
	_, exists := set.dict.Get(val)
	return exists
}

func (set *Set) Len() int {
	if set == nil || set.dict == nil {
		return 0
	}
	return set.dict.Len()
}

func (set *Set) ToSlice() []string {
	slice := make([]string, set.Len())
	i := 0
	set.dict.ForEach(func(key string, val interface{}) bool {
		if i < len(slice) {
			slice[i] = key
		} else {
			slice = append(slice, key)
		}
		i++
		return true
	})
	return slice
}

func (set *Set) ForEach(consumer func(member string) bool) {
	if set == nil || set.dict == nil {
		return
	}
	set.dict.ForEach(func(key string, val interface{}) bool {
		return consumer(key)
	})
}

func (set *Set) Copy() *Set {
	result := NewSet()
	set.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})
	return result
}

// Intersect 求多个集合的交集
func Intersect(sets ...*Set) *Set {
	result := NewSet()
	if len(sets) == 0 {
		return result
	}

	countMap := make(map[string]int)
	for _, set := range sets {
		set.ForEach(func(member string) bool {
			countMap[member]++
			return true
		})
	}
	for k, v := range countMap {
		if v == len(sets) { // 交集中的元素在每个集合中都存在
			result.Add(k)
		}
	}
	return result
}

// Union 求多个集合的并集
func Union(sets ...*Set) *Set {
	result := NewSet()
	for _, set := range sets {
		set.ForEach(func(member string) bool {
			result.Add(member)
			return true
		})
	}
	return result
}

// Diff 求多个集合的差集
func Diff(sets ...*Set) *Set {
	if len(sets) == 0 {
		return NewSet()
	}
	result := sets[0].Copy()
	for i := 1; i < len(sets); i++ {
		sets[i].ForEach(func(member string) bool {
			result.Remove(member)
			return true
		})
		if result.Len() == 0 {
			break
		}
	}
	return result
}

// RandomMembers 随机返回指定数量的key，可能包含重复的key
func (set *Set) RandomMembers(limit int) []string {
	if set == nil || set.dict == nil {
		return nil
	}
	return set.dict.RandomKeys(limit)
}

// RandomDistinctMembers 随机返回指定数量的key，不会包含重复的key
func (set *Set) RandomDistinctMembers(limit int) []string {
	return set.dict.RandomDistinctKeys(limit)
}
