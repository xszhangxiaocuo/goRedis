package list

import (
	"goRedis/interface/meta/list"
	"goRedis/lib/logger"
)

// LinkedList 双向链表
type LinkedList struct {
	first *node
	last  *node
	size  int
}

type node struct {
	val  any
	prev *node
	next *node
}

func NewLinkedList(vals ...any) *LinkedList {
	list := &LinkedList{}
	for _, v := range vals {
		list.Add(v)
	}
	return list
}

// Add 向链表尾部添加节点
func (list *LinkedList) Add(val any) {
	if list == nil {
		logger.Error("list is nil")
		return
	}
	n := &node{
		val: val,
	}
	if list.last == nil { // 空链表
		list.first = n
		list.last = n
	} else {
		n.prev = list.last
		list.last.next = n
		list.last = n
	}
	list.size++
}

func (list *LinkedList) find(index int) (n *node) {
	if index < list.size/2 {
		n = list.first
		for i := 0; i < index; i++ {
			n = n.next
		}
	} else {
		n = list.last
		for i := list.size - 1; i > index; i-- {
			n = n.prev
		}
	}
	return n
}

func (list *LinkedList) Get(index int) (val any) {
	if list == nil {
		logger.Error("list is nil")
		return nil
	}
	if index < 0 || index >= list.size {
		logger.Error("index out of bound")
		return nil
	}
	return list.find(index).val
}

func (list *LinkedList) Set(index int, val any) {
	if list == nil {
		logger.Error("list is nil")
		return
	}
	if index < 0 || index > list.size {
		logger.Error("index out of bound")
		return
	}
	n := list.find(index)
	n.val = val
}

func (list *LinkedList) Insert(index int, val any) {
	if list == nil {
		logger.Error("list is nil")
		return
	}
	if index < 0 || index > list.size {
		logger.Error("index out of bound")
		return
	}

	if index == list.size {
		list.Add(val)
		return
	}

	pivot := list.find(index)
	n := &node{
		val:  val,
		prev: pivot.prev,
		next: pivot,
	}
	if pivot.prev == nil {
		list.first = n
	} else {
		pivot.prev.next = n
	}
	pivot.prev = n
	list.size++
}

func (list *LinkedList) removeNode(n *node) {
	if n.prev == nil {
		list.first = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		list.last = n.prev
	} else {
		n.next.prev = n.prev
	}

	n.prev = nil
	n.next = nil

	list.size--
}

func (list *LinkedList) Remove(index int) (val any) {
	if list == nil {
		logger.Error("list is nil")
		return nil
	}
	if index < 0 || index > list.size {
		logger.Error("index out of bound")
		return nil
	}

	n := list.find(index)
	list.removeNode(n)
	return n.val
}

func (list *LinkedList) RemoveLast() (val any) {
	if list == nil {
		logger.Error("list is nil")
		return nil
	}
	if list.last == nil {
		return nil
	}
	n := list.last
	list.removeNode(n)
	return n.val
}

func (list *LinkedList) RemoveAllByVal(expected list.Expected) int {
	if list == nil {
		logger.Error("list is nil")
		return 0
	}
	n := list.first
	removed := 0
	var nextNode *node
	for n != nil {
		nextNode = n.next
		if expected(n.val) {
			list.removeNode(n)
			removed++
		}
		n = nextNode
	}
	return removed
}

// RemoveByVal removes at most `count` values of the specified value in this list
// scan from left to right
func (list *LinkedList) RemoveByVal(expected list.Expected, count int) int {
	if list == nil {
		logger.Error("list is nil")
		return 0
	}
	n := list.first
	removed := 0
	var nextNode *node
	for n != nil {
		nextNode = n.next
		if expected(n.val) {
			list.removeNode(n)
			removed++
		}
		if removed == count {
			break
		}
		n = nextNode
	}
	return removed
}

// ReverseRemoveByVal removes at most `count` values of the specified value in this list
// scan from right to left
func (list *LinkedList) ReverseRemoveByVal(expected list.Expected, count int) int {
	if list == nil {
		logger.Error("list is nil")
		return 0
	}
	n := list.last
	removed := 0
	var prevNode *node
	for n != nil {
		prevNode = n.prev
		if expected(n.val) {
			list.removeNode(n)
			removed++
		}
		if removed == count {
			break
		}
		n = prevNode
	}
	return removed
}

func (list *LinkedList) Len() int {
	if list == nil {
		logger.Error("list is nil")
		return 0
	}
	return list.size
}

func (list *LinkedList) ForEach(consumer list.Consumer) {
	if list == nil {
		logger.Error("list is nil")
		return
	}
	n := list.first
	i := 0
	for n != nil {
		goNext := consumer(i, n.val)
		if !goNext {
			break
		}
		i++
		n = n.next
	}
}

func (list *LinkedList) Contains(expected list.Expected) bool {
	contains := false
	list.ForEach(func(i int, actual any) bool {
		if expected(actual) {
			contains = true
			return false
		}
		return true
	})
	return contains
}

// Range returns elements which index within [start, stop)
func (list *LinkedList) Range(start int, stop int) []any {
	if list == nil {
		logger.Error("list is nil")
		return nil
	}
	if start < 0 || start >= list.size {
		logger.Error("`start` out of range")
		return nil
	}
	if stop < start || stop > list.size {
		logger.Error("`stop` out of range")
		return nil
	}

	sliceSize := stop - start
	slice := make([]any, sliceSize)
	n := list.first
	i := 0
	sliceIndex := 0
	for n != nil {
		if i >= start && i < stop {
			slice[sliceIndex] = n.val
			sliceIndex++
		} else if i >= stop {
			break
		}
		i++
		n = n.next
	}
	return slice
}
