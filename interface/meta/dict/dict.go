package dict

type Consumer func(key string, val any) bool

type Dict interface {
	Get(key string) (val any, exists bool)
	Len() int
	Put(key string, val any) (result int)         // 返回操作的键值对数量，插入返回1，更新返回0
	PutIfAbsent(key string, val any) (result int) // 如果key不存在则插入，插入返回1，不插入返回0
	PutIfExist(key string, val any) (result int)  // 如果key存在则插入，更新返回1，不更新返回0
	Remove(key string) (result int)               // 删除键值对，删除成功返回1，key不存在返回0
	Keys() []string                               // 返回所有的key
	ForEach(consumer Consumer)                    // 遍历所有的键值对
	RandomKeys(num int) []string                  // 随机返回num个key
	RandomDistinctKeys(num int) []string          // 随机返回num个不重复的key
	Clear()                                       // 清空字典
}
