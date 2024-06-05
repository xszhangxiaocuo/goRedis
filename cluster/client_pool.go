package cluster

import (
	"context"
	"errors"
	"fmt"
	pool "github.com/jolestar/go-commons-pool"
)

type connectionFactory struct {
	Peer string
}

func (f connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	var err error
	c, err := "假设有个client", nil
	if err != nil {
		return nil, err
	}
	//TODO 返回客户端连接，还未实现
	return pool.NewPooledObject(c), nil //返回客户端连接
}

func (f connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	o, ok := object.Object.(string)
	if !ok {
		return errors.New("类型转换出错")
	}
	fmt.Println(o) //todo 请实现关闭连接
	return nil
}

func (f connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (f connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (f connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
