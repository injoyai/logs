package logs

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*

	实现io.Writer接口


*/

//==============================Stdout==============================

type writeColor struct{ io.Writer }

func (this *writeColor) Color() bool { return true }

func NewWriteColor(writer io.Writer) io.Writer { return &writeColor{writer} }

//==============================Stdout==============================

type stdout struct {
	io.Writer
	Hook *hook
	once sync.Once
}

func (this *stdout) Write(p []byte) (int, error) {
	if this.Hook != nil {
		if this.Hook.Hook(p) {
			return this.Writer.Write(p)
		}
	}
	return len(p), nil
}

// Color 暂时用这个方法判断是否支持颜色
func (this *stdout) Color() bool { return true }

func (this *stdout) Input() {
	this.once.Do(func() {
		go func(s *stdout) {
			for {
				var input string
				//log.Println("请输入筛选条件(模糊搜索):")
				fmt.Scanln(&input)
				s.Hook.Like(input)
			}
		}(this)
	})
}

// newStdout 新建带Hook的标准输出
func newStdout() *stdout {
	s := &stdout{Writer: os.Stdout, Hook: newHook()}
	return s
}

//==============================消息总线==============================

// Trunk 消息总线
type trunk struct {
	c chan *HookMessage
}

func (this *trunk) Write(p []byte) (int, error) {
	select {
	case this.c <- &HookMessage{}:
	default:
	}
	return len(p), nil
}

func (this *trunk) Hook(e *Entity, msg string) {

}

func newTrunk(cap uint) *trunk {
	return &trunk{c: make(chan *HookMessage, cap)}
}

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

type httpClient struct {
	*http.Client
	method string
	url    string
	ch     *_chan
}

func (this *httpClient) Write(p []byte) (int, error) {
	return len(p), this.ch.Try(p)
}

func NewHTTPClient(method, url string) (io.Writer, error) {
	w := &httpClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: time.Second * 10,
		},
		method: method,
		url:    url,
		ch:     newChan(context.Background(), 100),
	}
	w.ch.handler = func(ctx context.Context, count int, data interface{}) {
		bs := data.([]byte)
		req, err := http.NewRequest(w.method, w.url, bytes.NewBuffer(bs))
		if err == nil {
			w.Client.Do(req)
		}
	}
	return w, nil
}

//==============================File==============================

// File 写入文件,自动打开关闭文件
type File struct {
	Name   string //实例名称
	Dir    string //路径
	Layout string //文件格式

	filename string         //现在打开的文件名称
	file     io.WriteCloser //文件流
}

func (this *File) Write(p []byte) (int, error) {

	//生成文件名
	layout := strings.ReplaceAll(this.Layout, "{name}", this.Name)
	filename := this.Dir + time.Now().Format(layout)

	//判断设置的文件地址是否有效
	if len(filename) == 0 {
		return 0, errors.New("无效文件地址")
	}

	//是否是新的文件,或文件流是否存在
	if filename != this.filename || this.file == nil {
		//获取文件不匹配
		//则关闭之前的文件,并重新生成文件
		if this.file != nil {
			if err := this.file.Close(); err != nil {
				return 0, err
			}
		}

		//判断文件是否存在,不存在则新建
		os.MkdirAll(filepath.Dir(filename), 0666)

		//新建文件(如果不存在),添加至文件最后
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModeAppend|os.ModePerm)
		if err != nil {
			log.Println(err)
			return 0, err
		}

		//赋值文件名和文件流,方便下次操作
		this.filename = filename
		this.file = file

	}

	//写入数据
	return this.file.Write(p)
}

// NewFile 写入文件
func NewFile(name, dir, layout string) io.Writer {
	f := &File{Name: name, Dir: dir, Layout: layout}
	return f
}

var DefaultRemoveFile = NewRemoveFile(0)

type RemoveFile struct {
	SaveTime   time.Duration       //保存时间
	RemoveFunc func(r *RemoveFile) //删除文件函数
}

func (this *RemoveFile) SetSaveTime(saveTime time.Duration) *RemoveFile {
	this.SaveTime = saveTime
	return this
}

func (this *RemoveFile) SetFunc(f func(r *RemoveFile)) *RemoveFile {
	this.RemoveFunc = f
	return this
}

func NewRemoveFile(saveTime time.Duration) *RemoveFile {
	r := &RemoveFile{
		SaveTime: saveTime,
		RemoveFunc: func(r *RemoveFile) {
			date := time.Now().Add(-r.SaveTime).Format("2006-01-02")
			files, _ := ioutil.ReadDir(DefaultDir)
			for _, v := range files {
				if len(v.Name()) == 10 && v.Name() <= date {
					_ = os.RemoveAll(DefaultDir + "/" + v.Name())
				}
			}
		},
	}
	go func(r *RemoveFile) {
		for {
			time.Sleep(time.Second * 5)
			if r.SaveTime > 0 && r.RemoveFunc != nil {
				r.RemoveFunc(r)
			}
			time.Sleep(time.Hour)
		}
	}(r)
	return r
}
