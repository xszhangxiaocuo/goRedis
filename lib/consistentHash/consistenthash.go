package consistentHash

import (
	"goRedis/lib/logger"
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunc func(data []byte) uint32 // 哈希函数

type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int          // 节点哈希（位置）列表。！因为要对节点的哈希值进行排序，sort函数默认不能实现排序。解决方法1.类型转换2.将uint32实现sort接口
	nodeHashMap map[int]string // string记录的节点名字、地址
	replicas    int            // 虚拟节点倍数
}

func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

func NewNodeMap(replicas int, hf HashFunc) *NodeMap {
	if hf == nil {
		hf = crc32.ChecksumIEEE // 默认哈希函数
	}

	return &NodeMap{ //部分初始化
		hashFunc:    hf,
		nodeHashMap: make(map[int]string),
		replicas:    replicas,
	}
}

func (m *NodeMap) AddNode(keys ...string) { // 传入名称或地址。将节点加入到列表中并重新排序
	for _, key := range keys {
		if key == "" {
			continue
		}
		for i := 0; i < m.replicas; i++ { //添加虚拟节点
			hash := int(m.hashFunc([]byte(key + strconv.Itoa(i)))) // 根据NodeMap自身维护的哈希函数对key哈希，得到哈希值。
			m.nodeHashs = append(m.nodeHashs, hash)                // 将计算出的哈希值添加到节点list中
			m.nodeHashMap[hash] = key                              // 将hash与真实的节点地址进行映射
		}
	}

	sort.Ints(m.nodeHashs) //将节点的哈希值进行排序
}

// RemoveNode 删除节点
func (m *NodeMap) RemoveNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		for i := 0; i < m.replicas; i++ { //删除虚拟节点
			hash := int(m.hashFunc([]byte(key + strconv.Itoa(i))))
			// 二分查找删除key在hash环中的下标
			idx := sort.SearchInts(m.nodeHashs, hash)
			if idx < len(m.nodeHashs) && m.nodeHashs[idx] == hash {
				m.nodeHashs = append(m.nodeHashs[:idx], m.nodeHashs[idx+1:]...) //删除hash环中的节点
			}
			delete(m.nodeHashMap, hash) //删除虚拟节点与真实节点的映射
		}
	}
}

func (m *NodeMap) PickNode(key string) string { //返回string是目标节点的地址或名称
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	nodeIDX := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash //找到大于该hash的第一个哈希，也就是找到了节点
	}) //返回满足（条件函数）的第一个下标
	if nodeIDX == len(m.nodeHashs) {
		nodeIDX = 0
	}
	logger.Info("PickNode: ", key, " -> ", m.nodeHashMap[m.nodeHashs[nodeIDX]], " nodeIDX: ", nodeIDX, " hash: ", hash, " nodeHashs: ", m.nodeHashs, " nodeHashMap: ", m.nodeHashMap) //debug
	return m.nodeHashMap[m.nodeHashs[nodeIDX]]
}
