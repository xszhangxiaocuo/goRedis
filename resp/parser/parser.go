// Package parser 协议解析器
package parser

import (
	"bufio"
	"errors"
	"goRedis/interface/resp"
	"goRedis/lib/logger"
	"goRedis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// Payload 对RESP协议字符串解析后的结构体
type Payload struct {
	Data resp.Reply
	Err  error
}

// readState 解析器的状态
type readState struct {
	readingMultiLine  bool     //解析的参数是否为多行
	expectedArgsCount int      //解析的参数个数
	msgType           byte     //消息类型
	args              [][]byte //解析出的参数数据
	bulkLen           int64    //当前需要读取的字符串长度(用于解析'$'开头的bulk string)
}

// finished 判断解析是否结束
func (r *readState) finished() bool {
	//需要的参数大于0并且已经解析完成的数据个数等于需要解析的参数个数
	return r.expectedArgsCount > 0 && len(r.args) == r.expectedArgsCount
}

// ParseStream 对上层暴露的解析器接口，异步并发地执行解析器，通过channel传递解析结果
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader)
	var err error
	var state readState
	var msg []byte

	for {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { //io错误，该进程结束
				ch <- &Payload{
					Err: err,
				}
				close(ch) //关闭channel
				return
			}
			//非io错误，继续读取
			ch <- &Payload{
				Err: err,
			}
			state = readState{} //清空解析器状态
			continue
		}
		//判断是否是多行解析模式
		if !state.readingMultiLine {
			if msg[0] == '*' { //数组
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: err,
					}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 { //*0\r\n 返回空数组
					ch <- &Payload{
						Data: reply.NewEmptyMultiBulkReply(),
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' { //字符串
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: err,
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 { //$-1\r\n返回null
					ch <- &Payload{
						Data: reply.NewNullBulkReply(),
					}
					state = readState{}
					continue
				}
			} else { //非多行模式，msg的type也不是数组和字符串，即单行回复
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else { //读取数组和字符串中的数据
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error:" + string(msg)),
				}
				state = readState{}
				continue
			}
			if state.finished() { //数据全部读取完毕
				var result resp.Reply
				if state.msgType == '*' { //数组
					result = reply.NewMultiBulkReply(state.args)
				} else if state.msgType == '$' { //字符串
					result = reply.NewBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}

		}
	}

}

// readLine 读取一行，返回（解析出的数据，是否为io错误，错误信息）
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error

	if state.bulkLen == 0 { //说明当前要读取的不是字符串，直接根据\r\n进行切分
		msg, err = bufReader.ReadBytes('\n')
		logger.Info("msg:" + string(msg))
		if err != nil { //io错误
			logger.Error(err)
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' { //非io错误，读取到的数据为空或结尾不以'\r\n'结尾，协议格式错误
			logger.Warn("protocol error:" + string(msg))
			return nil, false, errors.New("protocol error:" + string(msg))
		}
	} else { //当前要读取的是字符串，不能根据\r\n切分，严格按照bulkLen的大小进行读取
		msg = make([]byte, state.bulkLen+2)  //+2是因为要读取\r\n
		_, err = io.ReadFull(bufReader, msg) //将bufReader中的数据全部塞到msg中
		logger.Info("msg:" + string(msg))
		if err != nil { //io错误
			logger.Error(err)
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' { //非io错误，读取到的数据为空或结尾不以'\r\n'结尾，协议格式错误
			logger.Warn("protocol error:" + string(msg))
			return nil, false, errors.New("protocol error:" + string(msg))
		}
		state.bulkLen = 0 //清空需要读取的长度
	}
	return msg, false, nil
}

// parseMultiBulkHeader 初始化数组解析器，示例：*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n，传入*3\r\n进行初始化
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine int64
	expectedLine, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 32) //获取期望解析的参数个数
	if err != nil {
		logger.Warn(err)
		return errors.New("protocol error:" + string(msg))
	}

	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.readingMultiLine = true                //正在解析多行数据
		state.msgType = msg[0]                       //协议中的第一个字节标识了消息类型
		state.expectedArgsCount = int(expectedLine)  //期望解析的参数个数
		state.args = make([][]byte, 0, expectedLine) //存放解析出的数据
		return nil
	} else {
		logger.Warn(err)
		return errors.New("protocol error:" + string(msg))
	}
}

// parseBulkHeader 初始化字符串解析器，示例：$4\r\nPING\r\n，传入$4\r\n进行初始化
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 32) //获取字符串长度
	if err != nil {
		logger.Warn(err)
		return errors.New("protocol error:" + string(msg))
	}

	if state.bulkLen == -1 { //字符串为null
		return nil
	} else if state.bulkLen > 0 { //当作长度为1的数组处理
		state.readingMultiLine = true     //正在解析多行数据
		state.msgType = msg[0]            //协议中的第一个字节标识了消息类型
		state.expectedArgsCount = 1       //期望解析的参数个数
		state.args = make([][]byte, 0, 1) //存放解析出的数据
		return nil
	} else {
		logger.Warn(err)
		return errors.New("protocol error:" + string(msg))
	}
}

// parseSingleLineReply 解析单行回复，示例：+OK\r\n	-err\r\n	:3\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	msg0 := strings.TrimSuffix(string(msg), "\r\n") //删掉后缀\r\n
	var result resp.Reply

	switch msg0[0] { //判断回复类型
	case '+':
		result = reply.NewStatusReply(msg0[1:])
	case '-':
		result = reply.NewStandardErrReply(msg0[1:])
	case ':':
		num, err := strconv.ParseInt(msg0[1:], 10, 32)
		if err != nil {
			logger.Warn(err)
			return nil, errors.New("protocol error:" + msg0)
		}
		result = reply.NewIntReply(num)
	}
	return result, nil
}

func readBody(msg []byte, state *readState) error {
	var err error
	line := msg[:len(msg)-2] //删除末尾的\r\n

	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 32) //读取字符串长度
		if err != nil {
			logger.Warn(err)
			return errors.New("protocol error:" + string(msg))
		}
		if state.bulkLen <= 0 { //要读取的字符串长度小于等于0，设置结果为空
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
