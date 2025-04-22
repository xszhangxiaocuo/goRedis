package cluster

import (
	"fmt"
	"goRedis/interface/resp"
	"goRedis/lib/logger"
	"goRedis/resp/reply"
)

func makeRouter() map[string]CmdFunc {
	return map[string]CmdFunc{
		"ping":    local,
		"hello":   local,
		"exists":  defaultFunc,
		"type":    defaultFunc,
		"set":     defaultFunc,
		"setnx":   defaultFunc,
		"get":     defaultFunc,
		"getset":  defaultFunc,
		"select":  selectDB,
		"del":     del,
		"rename":  rename,
		"renamnx": rename,
		"flushdb": flushdb,
		"addnode": addNode,
	}
}

// 默认采用转发模式
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key) // 选择节点
	return cluster.relay(peer, c, args)
}

// 直接本地执行
func local(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(c, args)
}

// RENAME和RENAMNX，如果修改的key在同一个节点上，则直接修改，如果不在同一个节点上，则不允许跨节点修改
func rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.NewArgNumErrReply(string(args[0])) // 参数个数错误
	}
	src := string(args[1])
	dst := string(args[2])

	srcPeer := cluster.peerPicker.PickNode(src)
	dstPeer := cluster.peerPicker.PickNode(dst)

	if srcPeer == dstPeer { // 在同一个节点上，直接转发
		return cluster.relay(srcPeer, c, args)
	}

	return reply.NewStandardErrReply("ERR rename must within one node")
}

// FLUSHDB，删除当前数据库的所有key,需要群发
func flushdb(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcast(c, args)
	var errReply resp.ErrorReply
	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(resp.ErrorReply) // 只要有一个节点返回错误，就返回错误
			break
		}
	}
	if errReply != nil {
		return reply.NewStandardErrReply(errReply.Error())
	}
	return reply.NewOkReply()
}

func del(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.NewArgNumErrReply(string(args[0])) // 参数个数错误
	}
	var deleted int64 = 0
	// 找出所有的key对应的节点，分别转发
	for _, key := range args[1:] {
		r := defaultFunc(cluster, c, [][]byte{[]byte("DEL"), key})
		intReply, ok := r.(*reply.IntReply)
		if !ok {
			logger.Error(fmt.Sprintf("del '%s' error: %s", key, r.ToBytes()))
			continue
		}
		deleted += intReply.Code // 累加删除的个数
	}

	return reply.NewIntReply(deleted) // 返回删除的个数
}

func selectDB(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(c, args)
}

func addNode(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.NewArgNumErrReply(string(args[0])) // 参数个数错误
	}
	peers := args[1:]
	allNodes := cluster.GetNodes()
	if string(peers[len(peers)-1]) == "broadcast" {
		peers = peers[:len(peers)-1]
	}
	for _, arg := range peers { // 筛选出需要添加的节点
		peer := string(arg)
		if ok := cluster.NodeIsExist(peer); ok { // 如果已经存在该节点，则不添加
			continue
		}
		cluster.AddPeer(peer)
		allNodes[peer] = nil
		logger.Info(fmt.Sprintf("add node: %s", peer))
	}

	// 说明该节点是客户端直接连接的节点，需要广播这次添加节点操作
	if string(args[len(args)-1]) != "broadcast" {
		broadcastArgs := [][]byte{[]byte("addnode")}
		for n := range allNodes {
			broadcastArgs = append(broadcastArgs, []byte(n))
		}
		// 广播添加节点，通知所有节点，包括新添加的节点
		replies := cluster.broadcast0(c, append(broadcastArgs, []byte("broadcast")), allNodes)
		for n, r := range replies {
			if reply.IsErrReply(r) {
				logger.Error(fmt.Sprintf("add node error: node '%s': %s", n, r.ToBytes()))
				continue
			}
		}
	}
	return reply.NewOkReply()
}
