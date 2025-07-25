package logs

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"strings"
)

type Level uint8

const (
	LevelAll Level = iota
	LevelTrace
	LevelDebug
	LevelWrite
	LevelRead
	LevelInfo
	LevelWarn
	LevelError
	LevelNone Level = 255
)

func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "all":
		return LevelAll
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "write":
		return LevelWrite
	case "read":
		return LevelRead
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "err", "error":
		return LevelError
	case "none":
		return LevelNone
	default:
		return LevelAll
	}
}

// NewEntity 默认写到控制台和消息总线
func NewEntity(name string) *Entity {
	data := &Entity{
		Name:      name,
		Tag:       nil,
		Color:     0,
		ShowColor: true,
		caller:    0,
		Writer:    []io.Writer{Stdout},
		Formatter: DefaultFormatter,
		Level:     LevelAll,
		SelfLevel: LevelNone,
	}
	return data
}

// Entity [信息] 2020-01-02
type Entity struct {
	Name       string          //名称,例如 "INFO"
	Tag        []string        //标签.例如 "TCP"
	caller     int             //层级,默认3级
	callerBase int             //层级,基础
	Color      color.Attribute //颜色
	ShowColor  bool            //显示颜色
	Writer     []io.Writer     //输出
	Formatter  IFormatter      //格式
	Level      Level           //日志等级
	SelfLevel  Level           //自身日志等级
	Retry      int             //重试次数
}

// SetFormatter 设置格式化函数
func (this *Entity) SetFormatter(f IFormatter) *Entity {
	if f != nil {
		this.Formatter = f
	}
	return this
}

func (this *Entity) SetRetry(retry int) *Entity {
	this.Retry = retry
	return this
}

func (this *Entity) GetName() string {
	return this.Name
}

func (this *Entity) SetName(name string) *Entity {
	this.Name = name
	return this
}

// GetTag 获取tag
func (this *Entity) GetTag() []string {
	return this.Tag
}

// SetTag 设置tag
func (this *Entity) SetTag(s ...string) *Entity {
	this.Tag = s
	return this
}

// GetColor 获取颜色
func (this *Entity) GetColor() color.Attribute {
	return this.Color
}

// SetColor 设置颜色
func (this *Entity) SetColor(c color.Attribute) *Entity {
	this.Color = c
	return this
}

// SetShowColor 显示颜色
func (this *Entity) SetShowColor(b ...bool) *Entity {
	this.ShowColor = !(len(b) > 0 && !b[0])
	return this
}

func (this *Entity) SetLevel(level Level) *Entity {
	this.Level = level
	return this
}

func (this *Entity) SetSelfLevel(level Level) *Entity {
	this.SelfLevel = level
	return this
}

func (this *Entity) GetCaller() int {
	return this.caller + 4 + this.callerBase
}

// SetCaller 文件路径层级
func (this *Entity) SetCaller(n int) *Entity {
	this.caller = n
	return this
}

// setCaller 内置层级,不公开,为了解决默认函数层级多一级的情况
func (this *Entity) setCaller(n int) *Entity {
	this.callerBase = n
	return this
}

// SetWriter 设置输出,会覆盖之前设置的输出,并不会执行Close
func (this *Entity) SetWriter(writer ...io.Writer) *Entity {
	this.Writer = writer
	return this
}

// AddWriter 添加输出
func (this *Entity) AddWriter(writer ...io.Writer) *Entity {
	this.Writer = append(this.Writer, writer...)
	return this
}

// WriteToConsole 输出到控制台
func (this *Entity) WriteToConsole() *Entity {
	this.AddWriter(os.Stdout)
	return this
}

// WriteToFile 输出到文件 例"./output/logs/2006-01-02/{type}_15.log"
func (this *Entity) WriteToFile(filename string) *Entity {
	filename = strings.ReplaceAll(filename, "{type}", this.Name)
	this.AddWriter(NewFile(filename, 0))
	return this
}

// WriteToTrunk 写入到消息总线
func (this *Entity) WriteToTrunk() *Entity {
	this.AddWriter(Trunk)
	return this
}

// WriteToTCPClient 写入TCP客户端,color 是否传输颜色数据
func (this *Entity) WriteToTCPClient(addr string, color ...bool) error {
	writer, err := NewTCPClient(addr)
	if err != nil {
		return err
	}
	if len(color) > 0 && color[0] {
		writer = NewWriteColor(writer)
	}
	this.AddWriter(writer)
	return nil
}

// WriteToTCPServer 写入TCP服务器 ,color 是否传输颜色数据
func (this *Entity) WriteToTCPServer(port int, color ...bool) error {
	writer, err := NewTCPServer(port)
	if err != nil {
		return err
	}
	if len(color) > 0 && color[0] {
		writer = NewWriteColor(writer)
	}
	this.AddWriter(writer)
	return nil
}

// WriteToHTTPServer 写入HTTP服务器 ,color 是否传输颜色数据
func (this *Entity) WriteToHTTPServer(method, url string, color ...bool) error {
	writer := NewHTTPClient(method, url)
	if len(color) > 0 && color[0] {
		writer = NewWriteColor(writer)
	}
	this.AddWriter(writer)
	return nil
}

// Sprintf 格式化输出
func (this *Entity) Sprintf(format string, v ...interface{}) string {
	if this.Formatter == nil {
		this.Formatter = DefaultFormatter
	}
	return this.Formatter.Formatter(this, fmt.Sprintf(format, v...))
}

// Sprint 格式化输出
func (this *Entity) Sprint(v ...interface{}) string {
	if this.Formatter == nil {
		this.Formatter = DefaultFormatter
	}
	return this.Formatter.Formatter(this, fmt.Sprint(v...))
}

func (this *Entity) Sprintln(v ...interface{}) string {
	if this.Formatter == nil {
		this.Formatter = DefaultFormatter
	}
	return this.Formatter.Formatter(this, fmt.Sprintln(v...))
}

// Printf 格式化写入
func (this *Entity) Printf(format string, v ...interface{}) (int, error) {
	if this.Level > this.SelfLevel {
		return 0, nil
	}
	msg := []byte(this.Sprintf(format, v...))
	return this.Write(msg)
}

// Print 写入内容
func (this *Entity) Print(v ...interface{}) (int, error) {
	if this.Level > this.SelfLevel {
		return 0, nil
	}
	msg := []byte(this.Sprint(v...))
	return this.Write(msg)
}

// Println 写入内容,换行
func (this *Entity) Println(v ...interface{}) (int, error) {
	if this.Level > this.SelfLevel {
		return 0, nil
	}
	msg := []byte(this.Sprintln(v...))
	return this.Write(msg)
}

// Write 实现io.Writer
func (this *Entity) Write(p []byte) (n int, err error) {
	for _, w := range this.Writer {
		bs := p
		if w == nil {
			continue
		}
		if this.ShowColor && this.isColorWriter(w) {
			bs = []byte(color.New(this.Color).Sprint(string(p)))
		}
		for i := 0; i <= this.Retry; i++ {
			n, err = w.Write(bs)
			if err == nil {
				break
			}
		}
	}
	return
}

func (this *Entity) isColorWriter(w io.Writer) bool {
	if val, ok := w.(interface{ Color() bool }); ok {
		return val.Color()
	}
	return false
}
