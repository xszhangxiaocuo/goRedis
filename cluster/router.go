package cluster

import (
	"goRedis/interface/resp"
	"goRedis/resp/reply"
)

func makeRouter() map[string]CmdFunc {
	return map[string]CmdFunc{
		"hello":   defaultFunc,
		"exists":  defaultFunc,
		"type":    defaultFunc,
		"set":     defaultFunc,
		"setnx":   defaultFunc,
		"get":     defaultFunc,
		"getset":  defaultFunc,
		"select":  selectDB,
		"del":     del,
		"ping":    ping,
		"rename":  rename,
		"renamnx": rename,
		"flushdb": flushdb,
	}
}

// 默认采用转发模式
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key) // 选择节点
	return cluster.relay(peer, c, args)
}

// PING，直接本地执行
func ping(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
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
	replies := cluster.broadcast(c, args)
	var errReply resp.ErrorReply
	var deleted int64 = 0
	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(resp.ErrorReply) // 只要有一个节点返回错误，就返回错误
			break
		}
		intReply, ok := r.(*reply.IntReply)
		if !ok {
			errReply = reply.NewStandardErrReply("ERR type error")
			break
		}
		deleted += intReply.Code // 累加删除的个数
	}
	if errReply != nil {
		return reply.NewStandardErrReply(errReply.Error())
	}
	return reply.NewIntReply(deleted) // 返回删除的个数
}

func selectDB(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(c, args)
}
