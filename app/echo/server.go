package echo

import (
	"context"
	"fmt"
	"github.com/artisticbones/redis-learn/internal/pkg/tcp"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/**
 * TCP是字节流协议，不存在粘包问题
 * Redis 协议的指令、回复都是以消息为单位进行通信的
 */

type Server struct {
	addr     string `json:"addr"`
	port     uint   `json:"port"`
	protocol string `json:"protocol"`
}

func (s *Server) Network() string {
	return s.protocol
}

func (s *Server) String() string {
	return fmt.Sprintf("%s:%d", s.addr, s.port)
}

// ListenAndServe 监听并提供服务，并在收到 closeChan 发来的关闭通知后关闭
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// 监听关闭通知
	go func() {
		<-closeChan
		log.Println("shutting down...")
		// 停止监听，listener.Accept()会立即返回 io.EOF
		err := listener.Close()
		if err != nil {
			log.Fatalf("%#v\n", err)
		}
		// 关闭应用层服务器
		err = handler.Close()
		if err != nil {
			log.Fatalf("%#v\n", err)
		}
	}()

	// 异常退出后释放资源
	defer func() {
		err := listener.Close()
		if err != nil {
			log.Fatalf("%#v\n", err)
		}
		err = handler.Close()
		if err != nil {
			log.Fatalf("%#v\n", err)
		}
	}()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("%#v\n", err)
		}
		// 开启 goroutine 来处理新链接
		log.Println("accept link")
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(context.Background(), conn)
		}()
	}
	waitDone.Wait()
}

func ListenAndServeWithSignal(cfg *Server, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		switch <-sigCh {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen(cfg.Network(), cfg.String())
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("bind: %s, start listening...", cfg.String()))
	ListenAndServe(listener, handler, closeChan)
	return nil
}
