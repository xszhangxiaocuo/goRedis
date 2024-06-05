package consistentHash

import (
	"hash/crc32"
	"sort"
)

type HashFunc func(data []byte) uint32 //哈希函数

type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int          //节点哈希（位置）列表。！因为要对节点的哈希值进行排序，sort函数默认不能实现排序。解决方法1.类型转换2.将uint32实现sort接口
	nodeHashMap map[int]string //string记录的节点名字、地址
}

func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

func NewNodeMap(hf HashFunc) *NodeMap {
	if hf == nil {
		hf = crc32.ChecksumIEEE //默认哈希函数
	}

	return &NodeMap{ //部分初始化
		hashFunc:    hf,
		nodeHashMap: make(map[int]string),
	}
}

func (m *NodeMap) AddNode(keys ...string) { //传入名称或地址。将节点加入到列表中并重新排序
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))    //根据NodeMap自身维护的哈希函数对key哈希，得到哈希值。
		m.nodeHashs = append(m.nodeHashs, hash) //将计算出的哈希值添加到节点list中
		m.nodeHashMap[hash] = key               //记录key和hash的对应关系，以便根据hash得到key。
	}

	sort.Ints(m.nodeHashs) //将节点的哈希值进行排序
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
	return m.nodeHashMap[m.nodeHashs[nodeIDX]]
}
