package wait

import (
	"sync"
	"time"
)

//对sync.WaitGroup进行扩展，增加超时等待

// Wait 封装了sync.WaitGroup，但可以设置超时等待
type Wait struct {
	wg sync.WaitGroup // 内嵌 sync.WaitGroup
}

// Add 向 WaitGroup 计数器添加 delta，delta 可以是负数
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// Done 将 WaitGroup 计数器减一
func (w *Wait) Done() {
	w.wg.Done()
}

// Wait 阻塞，直到 WaitGroup 计数器为零
func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout 阻塞，直到 WaitGroup 计数器为零或超时
// 如果超时返回 true
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)  // 确保关闭通道
		w.Wait()        // 等待 WaitGroup 计数器为零
		c <- struct{}{} // 通知主线程完成
	}()
	select {
	case <-c:
		return false // 正常完成
	case <-time.After(timeout):
		return true // 超时
	}
}
