package cluster

import (
	"context"
	"errors"
	"goRedis/interface/resp"
	"goRedis/lib/utils"
	"goRedis/resp/client"
	"goRedis/resp/reply"
	"strconv"
)

func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("未找到连接")
	}
	ctx := context.Background()
	object, err := pool.BorrowObject(ctx)
	if err != nil {
		return nil, err
	}
	c, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("类型转换错误")
	}

	return c, err
}

func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("未找到连接")
	}
	return pool.ReturnObject(context.Background(), peerClient) // 将连接放回连接池

}

// 转发，connection是resp里面记录用户信息的conn
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	// 本地执行
	if peer == cluster.self {
		return cluster.db.Exec(c, args)
	}
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.NewStandardErrReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex()))) // 先选择DB
	return peerClient.Send(args)
}

// 群发广播
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		result := cluster.relay(node, c, args) //调用转发函数
		results[node] = result
	}
	return results
}
