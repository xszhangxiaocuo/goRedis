package list

import (
	"container/list"
	list2 "goRedis/interface/meta/list"
	"goRedis/lib/logger"
)

// pageSize 每页大小必须是偶数，因为在插入满页时会将页分为两半
const pageSize = 1024

// QuickList 封装了一个双向链表，每个节点是一个 page ， page 的大小是 pageSize
type QuickList struct {
	data *list.List
	size int
}

// iterator 迭代器，用于在 [-1, ql.Len()] 之间移动
type iterator struct {
	node   *list.Element
	offset int
	ql     *QuickList
}

func NewQuickList() *QuickList {
	l := &QuickList{
		data: list.New(),
	}
	return l
}

// get 返回当前迭代器指向的 page 中偏移量为 offset 的元素
func (iter *iterator) get() any {
	return iter.page()[iter.offset]
}

// set 设置当前迭代器指向的 page 中偏移量为 offset 的元素
func (iter *iterator) set(val any) {
	page := iter.page()
	page[iter.offset] = val
}

// page 返回当前迭代器指向的 page ，即当前迭代器中的 node
func (iter *iterator) page() []any {
	return iter.node.Value.([]any)
}

// next 返回 offset 是否在范围内
func (iter *iterator) next() bool {
	page := iter.page()
	if iter.offset < len(page)-1 {
		iter.offset++
		return true
	}
	// 最后一页
	if iter.node == iter.ql.data.Back() {
		iter.offset = len(page)
		return false
	}
	// 移动到下一页
	iter.offset = 0
	iter.node = iter.node.Next()
	return true
}

// remove 移除当前迭代器指向的元素
func (iter *iterator) remove() any {
	page := iter.page()
	val := page[iter.offset]
	page = append(page[:iter.offset], page[iter.offset+1:]...) // 移除元素
	if len(page) > 0 {
		// page 不为空，更新 iter.node 和 iter.offset
		iter.node.Value = page
		if iter.offset == len(page) {
			// iter 在 page 的末尾，移动到下一页
			if iter.node != iter.ql.data.Back() {
				iter.node = iter.node.Next()
				iter.offset = 0
			}
		}
	} else {
		// page 为空，更新 iter.node 和 iter.offset
		if iter.node == iter.ql.data.Back() {
			if prevNode := iter.node.Prev(); prevNode != nil {
				iter.ql.data.Remove(iter.node)
				iter.node = prevNode
				iter.offset = len(prevNode.Value.([]any))
			} else {
				// 移除了最后一页，iter 指向 nil
				iter.ql.data.Remove(iter.node)
				iter.node = nil
				iter.offset = 0
			}
		} else { // 移动到下一页
			nextNode := iter.node.Next()
			iter.ql.data.Remove(iter.node)
			iter.node = nextNode
			iter.offset = 0
		}
	}
	iter.ql.size--
	return val
}

// prev 返回 offset 是否在范围内
func (iter *iterator) prev() bool {
	if iter.offset > 0 {
		iter.offset--
		return true
	}
	// 第一页
	if iter.node == iter.ql.data.Front() {
		iter.offset = -1
		return false
	}
	// 移动到上一页
	iter.node = iter.node.Prev()
	prevPage := iter.node.Value.([]any)
	iter.offset = len(prevPage) - 1
	return true
}

// atEnd 返回是否到达尾部
func (iter *iterator) atEnd() bool {
	if iter.ql.data.Len() == 0 {
		return true
	}
	if iter.node != iter.ql.data.Back() {
		return false
	}
	page := iter.page()
	return iter.offset == len(page)
}

// atBegin 返回是否到达头部
func (iter *iterator) atBegin() bool {
	if iter.ql.data.Len() == 0 {
		return true
	}
	if iter.node != iter.ql.data.Front() {
		return false
	}
	return iter.offset == -1
}

func (ql *QuickList) Add(val any) {
	ql.size++
	if ql.data.Len() == 0 {
		page := make([]any, 0, pageSize)
		page = append(page, val)
		ql.data.PushBack(page)
		return
	}
	backNode := ql.data.Back()
	backPage := backNode.Value.([]any)
	if len(backPage) == cap(backPage) { // 一个 page 已满
		page := make([]any, 0, pageSize)
		page = append(page, val)
		ql.data.PushBack(page)
		return
	}
	backPage = append(backPage, val)
	backNode.Value = backPage
}

// find 返回给定索引的 page 和在 page 中的偏移量
func (ql *QuickList) find(index int) *iterator {
	if ql == nil {
		logger.Error("list is nil")
	}
	if index < 0 || index >= ql.size {
		logger.Error("index out of bound")
	}
	var n *list.Element
	var page []any
	var pageBeg int
	if index < ql.size/2 { // 从头部开始查找
		n = ql.data.Front()
		pageBeg = 0
		for {
			page = n.Value.([]any)
			if pageBeg+len(page) > index {
				break
			}
			pageBeg += len(page)
			n = n.Next()
		}
	} else { // 从尾部开始查找
		n = ql.data.Back()
		pageBeg = ql.size
		for {
			page = n.Value.([]any)
			pageBeg -= len(page)
			if pageBeg <= index {
				break
			}
			n = n.Prev()
		}
	}
	pageOffset := index - pageBeg
	return &iterator{
		node:   n,
		offset: pageOffset,
		ql:     ql,
	}
}

func (ql *QuickList) Get(index int) (val any) {
	iter := ql.find(index)
	return iter.get()
}

func (ql *QuickList) Set(index int, val any) {
	iter := ql.find(index)
	iter.set(val)
}

func (ql *QuickList) Insert(index int, val any) {
	if index == ql.size {
		ql.Add(val)
		return
	}
	iter := ql.find(index)
	page := iter.node.Value.([]any)
	if len(page) < pageSize { // 判断当前页是否已满
		page = append(page[:iter.offset+1], page[iter.offset:]...)
		page[iter.offset] = val
		iter.node.Value = page
		ql.size++
		return
	}
	// 插入到一个满页会导致内存拷贝，所以将一个满页分成两个半页
	var nextPage []any
	nextPage = append(nextPage, page[pageSize/2:]...) // pageSize 必须是偶数
	page = page[:pageSize/2]
	if iter.offset < len(page) { // 前半页
		page = append(page[:iter.offset+1], page[iter.offset:]...)
		page[iter.offset] = val
	} else { // 后半页
		i := iter.offset - pageSize/2
		nextPage = append(nextPage[:i+1], nextPage[i:]...)
		nextPage[i] = val
	}
	// 保存当前页和下一页
	iter.node.Value = page
	ql.data.InsertAfter(nextPage, iter.node) // 在当前页后插入下一页
	ql.size++
}

func (ql *QuickList) Remove(index int) any {
	iter := ql.find(index)
	return iter.remove()
}

func (ql *QuickList) Len() int {
	return ql.size
}

// RemoveLast 移除最后一个元素
func (ql *QuickList) RemoveLast() any {
	if ql.Len() == 0 {
		return nil
	}
	ql.size--
	lastNode := ql.data.Back()
	lastPage := lastNode.Value.([]any)
	if len(lastPage) == 1 {
		ql.data.Remove(lastNode)
		return lastPage[0]
	}
	val := lastPage[len(lastPage)-1]
	lastPage = lastPage[:len(lastPage)-1]
	lastNode.Value = lastPage
	return val
}

// RemoveAllByVal 移除所有符合预期的元素
func (ql *QuickList) RemoveAllByVal(expected list2.Expected) int {
	iter := ql.find(0)
	removed := 0
	for !iter.atEnd() {
		if expected(iter.get()) {
			iter.remove()
			removed++
		} else {
			iter.next()
		}
	}
	return removed
}

func (ql *QuickList) RemoveByVal(expected list2.Expected, count int) int {
	if ql.size == 0 {
		return 0
	}
	iter := ql.find(0)
	removed := 0
	for !iter.atEnd() {
		if expected(iter.get()) {
			iter.remove()
			removed++
			if removed == count {
				break
			}
		} else {
			iter.next()
		}
	}
	return removed
}

func (ql *QuickList) ReverseRemoveByVal(expected list2.Expected, count int) int {
	if ql.size == 0 {
		return 0
	}
	iter := ql.find(ql.size - 1)
	removed := 0
	for !iter.atBegin() {
		if expected(iter.get()) {
			iter.remove()
			removed++
			if removed == count {
				break
			}
		}
		iter.prev()
	}
	return removed
}

func (ql *QuickList) ForEach(consumer list2.Consumer) {
	if ql == nil {
		logger.Error("list is nil")
	}
	if ql.Len() == 0 {
		return
	}
	iter := ql.find(0)
	i := 0
	for {
		goNext := consumer(i, iter.get())
		if !goNext {
			break
		}
		i++
		if !iter.next() {
			break
		}
	}
}

func (ql *QuickList) Contains(expected list2.Expected) bool {
	contains := false
	ql.ForEach(func(i int, actual any) bool {
		if expected(actual) {
			contains = true
			return false
		}
		return true
	})
	return contains
}

// Range 遍历范围 [start, stop)，使用迭代器实现
func (ql *QuickList) Range(start int, stop int) []any {
	if start < 0 || start >= ql.Len() {
		logger.Error("`start` out of range")
	}
	if stop < start || stop > ql.Len() {
		logger.Error("`stop` out of range")
	}
	sliceSize := stop - start
	slice := make([]any, 0, sliceSize)
	iter := ql.find(start)
	i := 0
	for i < sliceSize {
		slice = append(slice, iter.get())
		iter.next()
		i++
	}
	return slice
}
