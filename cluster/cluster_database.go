// 集群数据库：在这层做请求转发。底层的单机数据库为standAlone_database
package cluster

import (
	"context"
	pool "github.com/jolestar/go-commons-pool"
	"goRedis/config"
	database2 "goRedis/database"
	"goRedis/interface/database"
	"goRedis/interface/resp"
	"goRedis/lib/consistentHash"
	"goRedis/lib/logger"
	"goRedis/resp/reply"
	"strings"
)

type ClusterDatabase struct { //Cluster节点:A要维护一组对B、一组对C节点的客户端。并发获取多个连接而不是一个连接。
	self           string
	nodes          []string                    //记录集群中所有的节点
	peerPicker     *consistentHash.NodeMap     //一致性哈希管理器，可以判空、添加节点、选择节点
	peerConnection map[string]*pool.ObjectPool //连接池,每个cluster节点都需要多个链接，用连接池维护
	db             database.Database           //底层的单机数据库
}

func NewClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             database2.NewStandaloneDataBase(),
		peerPicker:     consistentHash.NewNodeMap(config.Properties.ClusterReplicas, nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.nodes = nodes
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	// 为每个节点创建连接池
	for _, peer := range config.Properties.Peers {
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}

	return cluster
}

type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply

var router = makeRouter()

func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(r)
			result = reply.NewUnknownErrReply()
		}
	}()
	cmd := strings.ToLower(string(args[0]))
	if cmdFunc, ok := router[cmd]; !ok {
		result = reply.NewStandardErrReply("ERR not supported command")
	} else {
		result = cmdFunc(c, client, args)
	}
	return
}

func (c *ClusterDatabase) Close() error {
	return c.db.Close()
}

func (c *ClusterDatabase) AfterClientClose(client resp.Connection) error {
	return c.db.AfterClientClose(client)
}
