package aof

import (
	"fmt"
	"testing"
)

func TestAof(t *testing.T) {
	a, err := NewAofHandler(nil)
	if err != nil {
		fmt.Println("初始化AofHandler失败")
	}
	a.AddAof(1)

}
