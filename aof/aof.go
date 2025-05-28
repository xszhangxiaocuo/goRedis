package aof

import (
	"goRedis/config"
	"goRedis/interface/database"
	"goRedis/lib/logger"
	"goRedis/lib/utils"
	"goRedis/resp/connection"
	"goRedis/resp/parser"
	"goRedis/resp/reply"
	"io"
	"os"
	"strconv"
)

const aofBufferSize = 1 << 16

type payload struct {
	cmdLine database.CmdLine
	dbIndex int
}

type AofHandler struct {
	database    database.Database
	aofChan     chan *payload //存储引擎写操作时，传递消息
	aofFile     *os.File
	aofFileName string
	currentDB   int //维护当前库的id
}

func NewAofHandler(database database.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFileName = config.Properties.AppendFilename
	handler.database = database
	handler.LoadAof() // 当调用NewAofHandler时，是启动操作。先把写在硬盘上的aof文件恢复到内存中来。

	aofFile, err := os.OpenFile(handler.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofBufferSize)
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// ↓异步落盘\持久化
func (handler *AofHandler) AddAof(dbIndex int, cmd database.CmdLine) { //传入：几号DB数据库
	if config.Properties.AppendOnly && handler.aofChan != nil {
		//新建pyload
		handler.aofChan <- &payload{ //将传入参数组装为payload并传到channel
			cmdLine: cmd,
			dbIndex: dbIndex,
		}
	}
}

// 接收aofChan中的payload
func (handler *AofHandler) handleAof() {
	handler.currentDB = 0

	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB { //检查是否跟上一个DB一样,如果不一样，插入select语句
			args := utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))
			data := reply.NewMultiBulkReply(args).ToBytes() //得到写入文件的字节
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Error(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.NewMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Error(err)
		}
	}
}

// loadAof
func (handler *AofHandler) LoadAof() {
	file, err := os.Open(handler.aofFileName)
	if err != nil {
		logger.Error(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)
	tempConnection := &connection.RESPConn{}
	for p := range ch { // 接收解析器的返回值
		if p.Err != nil {
			if p.Err == io.EOF {
				logger.Info("AOF: load aof file finished")
				break
			}
			logger.Error(p.Err)
			continue
		}
		if p.Data == nil {
			logger.Error("AOF: empty payload")
		}
		multiBulkReply, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("AOF: exec error: ", p.Data)
			continue
		}
		r := handler.database.Exec(tempConnection, multiBulkReply.Args) //执行命令
		if reply.IsErrReply(r) {
			logger.Error("AOF: exec error: ", r.ToBytes())
			continue
		}
	}
}
