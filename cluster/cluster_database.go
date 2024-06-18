package cluster //集群数据库：在这层做请求转发。底层的单机数据库为standAlone_database

import (
	"context"
	pool "github.com/jolestar/go-commons-pool"
	"goRedis/config"
	"goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/lib/consistentHash"
)

type ClusterDatabase struct { //Cluster节点:A要维护一组对B、一组对C节点的客户端。并发获取多个连接而不是一个连接。
	self           string
	nodes          []string                    //记录集群中所有的节点
	peerPicker     *consistentHash.NodeMap     //一致性哈希管理器，可以判空、添加节点、选择节点
	peerConnection map[string]*pool.ObjectPool //连接池                         //ConnectionPool//每个cluster节点都需要多个链接，用连接池维护
	db             database.Database           //底层的单机数据库
}

func NewClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             nil,                            //todo NewStandAloneDatabase()未实现
		peerPicker:     consistentHash.NewNodeMap(nil), //传入哈希函数
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	ctx := context.Background()
	for _, node := range nodes {
		cluster.peerPicker.AddNode(node)
		cluster.peerConnection[node] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: node,
		})
		//
	}
	cluster.nodes = nodes

	return cluster
}

func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	//TODO implement me
	panic("implement me")
}

func (c *ClusterDatabase) Close() error {
	//TODO implement me
	panic("implement me")
}

func (c *ClusterDatabase) AfterClientClose(client resp.Connection) error {
	//TODO implement me
	panic("implement me")
}
