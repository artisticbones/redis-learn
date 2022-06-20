package proto

import (
	"bufio"
	"errors"
	"github.com/artisticbones/redis-learn/internal/redis"
	"io"
	"log"
	"runtime/debug"
)

/**
 * RESP 是一个二进制安全的文本协议，工作于 TCP 协议上。
 * RESP 以行作为单位，客户端和服务器发送的命令或数据一律以 \r\n （CRLF）作为换行符。
 * RESP 的二进制安全性允许我们在 key 或者 value 中包含 \r 或者 \n 这样的特殊字符。
 * 在使用 redis 存储 protobuf、msgpack 等二进制数据时，二进制安全性尤为重要。
 * RESP 定义了5种格式：
 ** 简单字符串(Simple String): 服务器用来返回简单的结果，比如"OK"。非二进制安全，且不允许换行。					|	以“+”开始，如："+OK\r\n"
 ** 错误信息(Error): 服务器用来返回简单的错误信息，比如"ERR Invalid Synatx"。非二进制安全，且不允许换行。		|	以"-"开始，如："-ERR Invalid Synatx\r\n"
 ** 整数(Integer): llen、scard 等命令的返回值, 64位有符号整数。											|	以":"开始，如：":1\r\n"
 ** 字符串(Bulk String): 二进制安全字符串, 比如 get 等命令的返回值。										|	以"$"开始
 ** 数组(Array, 又称 Multi Bulk Strings): Bulk String 数组，客户端发送指令以及 lrange 等命令响应的格式。	|	以"*"开始
 */

type Payload struct {
	Data redis.Reply
	Err  error
}

// ParseStream 通过 io.Reader 读取数据并将结果通过 channel 将结果返回给调用者
// 流式处理的接口适合供客户端/服务端使用
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

type readState struct {
	readingMultiLine  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulkLen           int64
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf(string(debug.Stack()))
		}
	}()
	// 初始化读取状态
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for true {
		// read line
		// 上文中我们提到 RESP 是以行为单位的
		// 因为行分为简单字符串和二进制安全的BulkString，我们需要封装一个 readLine 函数来兼容
		msg, err = readLine(bufReader, &state)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			ch <- &Payload{
				Err: err,
			}
			close(ch)
			return
		}
		if err != nil {
			// protocol err, reset read state
			ch <- &Payload{
				Err: err,
			}
			state = readState{}
			continue
		}
		// 接下来我们对刚刚读取的行进行解析
		// 我们简单的将 Reply 分为两类:
		// 单行: StatusReply, IntReply, ErrorReply
		// 多行: BulkReply, MultiBulkReply
		if !state.readingMultiLine {
			if msg[0] == '*' {
				err = parseMultiBulkHeader(msg, &state)
			}
		}
	}
}

func parseMultiBulkHeader(msg []byte, r *readState) error {
	return nil
}

func readLine(bufReader *bufio.Reader, state *readState) ([]byte, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { // read normal line
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, errors.New("protocol error: " + string(msg))
		}
	} else { // read bulk line (binary safe)
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
			return nil, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, nil
}
