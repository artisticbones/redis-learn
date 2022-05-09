package proto

import (
	"github.com/artisticbones/redis-learn/internal/redis"
	"io"
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

func parse0(reader io.Reader, ch chan<- *Payload) {

}
