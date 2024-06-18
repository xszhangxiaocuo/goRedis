package cluster

import (
	"context"
	"errors"
	"goRedis/interface/resp"
	"goRedis/resp/reply"
	"log"
)

func (cluster *ClusterDatabase) getPeerClient(peer string) (any, error) { //todo 返回redis服务器连接

	ctx := context.Background()
	pool := cluster.peerConnection[peer]
	object, err := pool.BorrowObject(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c, ok := object.(*string) //todo：把它转换为redis服务器连接
	if !ok {
		return nil, errors.New("类型转换错误")
	}

	return c, err
}

func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient string) error { //todo 传入redis连接
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("未找到连接")
	}
	return pool.ReturnObject(context.Background(), peerClient)

}

func (cluster *ClusterDatabase) relay(peerIp string, c resp.Connection, args [][]byte) resp.Reply { //转发。connection是resp里面记录用户信息的conn
	if peerIp == cluster.self {
		return cluster.db.Exec(c, args)
	}
	client, err := cluster.getPeerClient(peerIp)
	if err != nil {
		return reply.NewStandardErrReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peerIp, client.(string))
	}()
	//todo peerClient.Send(utils.ToCmdLine("select",strconv.Itoa(c.getDBIndex())))
	return nil //todo 返回转发的响应return client.Send(args)
}

func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply { //[][]byte是指令
	results := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		result := cluster.relay(node, c, args) //调用转发函数
		results[node] = result
	}
	return results
}
