package tcp

import (
	"MyRedis/interface/tcp"
	"MyRedis/lib/logger"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// A tcp server

// Config stores tcp server properties
type Config struct {
	Address  string        `yaml:"address"`
	MaxCount uint32        `yaml:"max-connect"`
	Timeout  time.Duration `yaml:"timeout"`
}

// ListenAndServeWithSignal binds port and handle requests, blocking until receive stop signal
// 监听中断信号并通过 closeChan 通知服务器关闭
func ListenAndServerWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	//cfg.Address = listener.Addr().String()
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe binds port and handle requests, blocking until close
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// listen signal (监听关闭通知)
	go func() {
		<-closeChan // 无中断信号时,此处阻塞
		// 以下为中断信号来了之后的处理, 关闭listener和handler
		logger.Info("shutting down...")
		// 停止监听，listener.Accept()会立即返回 io.EOF
		_ = listener.Close() // listener.Accept() will return err immediately
		// 关闭应用层服务器
		_ = handler.Close() // close connection
	}()

	// listen port (在异常退出后释放资源, 即非中断信号引起的关闭)
	defer func() {
		// close during unexpected error
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		// 监听端口, 阻塞直到收到新连接或者出现错误
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		// handle -> 开启新的 goroutine 来处理连接
		logger.Info("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait() // 所有handler都关闭才退出
}
