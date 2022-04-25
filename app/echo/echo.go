package echo

import (
	"bufio"
	"context"
	"github.com/artisticbones/redis-learn/pkg/sync/atomic"
	"github.com/artisticbones/redis-learn/pkg/sync/wait"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// Client 客户端连接的抽象
type Client struct {
	// tcp connection
	conn    net.Conn
	Waiting wait.Wait
}

func (c *Client) SetConn(conn net.Conn) {
	c.conn = conn
}

func (c *Client) GetConn() net.Conn {
	return c.conn
}

type Handler struct {
	// 保存所有工作状态client的集合(把map当set用)
	// 需使用并发安全的容器
	activeConn sync.Map

	// 关闭转态识别位
	closing atomic.Boolean
}

func NewEchoHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	// 关闭中的 handler 不会处理新连接
	if h.closing.Get() {
		conn.Close()
		return
	}

	client := &Client{}
	client.SetConn(conn)
	h.activeConn.Store(client, struct{}{})
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
				h.activeConn.Delete(client)
			} else {
				log.Println(err)
			}
			return
		}
		// 发送数据前先置为waiting状态，阻止连接被关闭
		client.Waiting.Add(1)
		b := []byte(msg)
		conn.Write(b)
		// 发送完毕，结束 waiting
		client.Waiting.Done()
	}
}

func (c *Client) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	err := c.GetConn().Close()
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Close() error {
	log.Println("handler shutting down...")
	h.closing.Set(true)

	h.activeConn.Range(func(key, value interface{}) bool {
		client, ok := key.(*Client)
		if !ok {
			return false
		}
		err := client.Close()
		if err != nil {
			return false
		}
		return true
	})
	return nil
}
