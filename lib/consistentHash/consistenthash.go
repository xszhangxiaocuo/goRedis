package consistentHash

import "hash/crc32"

type HashFunc func(data []byte) uint32 //哈希函数

type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int          //节点哈希（位置）列表。！因为要对节点的哈希值进行排序，sort函数默认不能实现排序。解决方法1.类型转换2.将uint32实现sort接口
	nodeHashMap map[int]string //记录写入的时候计算的哈希结果最终该去哪个节点
}

func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

func NewNodeMap(hf HashFunc) *NodeMap {
	if hf == nil {
		hf = crc32.ChecksumIEEE //
	}

	return &NodeMap{ //部分初始化
		hashFunc:    hf,
		nodeHashMap: make(map[int]string),
	}
}

func (m *NodeMap) AddNode(keys ...string) { //可变长参数
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key))) //返回值为uint32，类型转换为int
		m.nodeHashs = append(m.nodeHashs, hash)
	}
}
