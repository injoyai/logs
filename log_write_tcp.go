package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

//==============================TCP==============================

// tcpClient tcp客户端
type tcpClient struct {
	net.Conn
	ch *_chan
}

func (this *tcpClient) Write(p []byte) (int, error) {
	return len(p), this.ch.Try(p)
}

// NewTCPClient 推送至指定TCP服务器,断线重连
func NewTCPClient(addr string) (io.Writer, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	t := &tcpClient{Conn: c, ch: newChan(context.Background(), 100)}
	t.ch.handler = func(ctx context.Context, count int, data interface{}) {
		p := data.([]byte)
		_, err := t.Conn.Write(p)
		if err != nil {
			c, err := net.Dial("tcp", t.Conn.RemoteAddr().String())
			if err == nil {
				t.Conn = c
			}
		}
	}
	return t, nil
}

// DialTCP 监听tcp数据
func DialTCP(addr string, dealFunc func(p []byte)) error {

	c, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	buf := bufio.NewReader(c)

	readAll := func(buf *bufio.Reader) (bytes []byte, err error) {
		num := 1 << 10
		for {
			data := make([]byte, num)
			length, err := buf.Read(data)
			if err != nil {
				return nil, err
			}
			bytes = append(bytes, data[:length]...)
			if length < num || buf.Buffered() == 0 {
				//缓存没有剩余的数据
				return bytes, err
			}
		}
	}
	go func() {
		defer func() {
			i := time.Second
			for {
				if DialTCP(addr, dealFunc) != nil {
					return
				}
				if i < time.Second*32 {
					i *= 2
				}
				<-time.After(i)
			}
		}()
		for {
			bytes, err := readAll(buf)
			if err != nil {
				return
			}
			dealFunc(bytes)
		}
	}()

	return nil
}

type tcpServer struct {
	listener net.Listener
	conn     map[string]net.Conn
	mu       sync.RWMutex
	ch       *_chan
}

func (this *tcpServer) run() {
	for {
		c, err := this.listener.Accept()
		if err != nil {
			return
		}
		this.mu.Lock()
		this.conn[c.RemoteAddr().String()] = c
		this.mu.Unlock()
	}
}

func (this *tcpServer) getConn() map[string]net.Conn {
	m := map[string]net.Conn{}
	this.mu.RLock()
	defer this.mu.RUnlock()
	for i, v := range this.conn {
		m[i] = v
	}
	return m
}

func (this *tcpServer) delConn(key ...string) {
	this.mu.Lock()
	for _, v := range key {
		delete(this.conn, v)
	}
	this.mu.Unlock()
}

func (this *tcpServer) Write(p []byte) (int, error) {
	return len(p), this.ch.Try(p)
}

// NewTCPServer 推送至TCP所有连接的客户端
func NewTCPServer(port int) (io.Writer, error) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	writer := &tcpServer{
		listener: listener,
		conn:     make(map[string]net.Conn),
		ch:       newChan(context.Background(), 100),
	}

	writer.ch.handler = func(ctx context.Context, count int, data interface{}) {
		p := data.([]byte)
		errKey := []string(nil)
		for i, v := range writer.getConn() {
			if _, err := v.Write(p); err != nil {
				errKey = append(errKey, i)
			}
		}
		writer.delConn(errKey...)
	}

	go writer.run()

	return writer, nil
}
