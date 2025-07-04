package logs

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"sync"
	"time"
)

const (
	DefaultDir    = "./output/logs/"
	DefaultLayout = "2006-01-02/{name}_15.log"
)

var (
	m              = sync.Map{}
	DefaultTrace   = NewEntity("追溯").SetSelfLevel(LevelTrace).setCaller(1).SetColor(color.FgGreen)
	DefaultWrite   = NewEntity("写入").SetSelfLevel(LevelWrite).setCaller(1).SetColor(color.FgBlue)
	DefaultRead    = NewEntity("读取").SetSelfLevel(LevelRead).setCaller(1).SetColor(color.FgBlue)
	DefaultInfo    = NewEntity("信息").SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgCyan)
	DefaultDebug   = NewEntity("调试").SetSelfLevel(LevelDebug).setCaller(1).SetColor(color.FgYellow)
	DefaultWarn    = NewEntity("警告").SetSelfLevel(LevelWarn).setCaller(1).SetColor(color.FgMagenta)
	DefaultErr     = NewEntity("错误").SetSelfLevel(LevelError).setCaller(1).SetColor(color.FgRed) //.WriteToFile(DefaultDir, DefaultLayout)
	DefaultRed     = NewEntity("").SetFormatter(FOriginal).SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgRed)
	DefaultYellow  = NewEntity("").SetFormatter(FOriginal).SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgYellow)
	DefaultBlue    = NewEntity("").SetFormatter(FOriginal).SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgBlue)
	DefaultGreen   = NewEntity("").SetFormatter(FOriginal).SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgGreen)
	DefaultCyan    = NewEntity("").SetFormatter(FOriginal).SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgCyan)
	DefaultMagenta = NewEntity("").SetFormatter(FOriginal).SetSelfLevel(LevelInfo).setCaller(1).SetColor(color.FgMagenta)

	// Trunk 消息总线,公共Writer,
	Trunk = newTrunk(1000)

	// Stdout 系统标准输出, NewWriteColor 支持颜色输出
	Stdout = newStdout()
)

func init() {
	color.NoColor = false
	m.Store(DefaultTrace.GetName(), DefaultTrace)
	m.Store(DefaultWrite.GetName(), DefaultWrite)
	m.Store(DefaultRead.GetName(), DefaultRead)
	m.Store(DefaultInfo.GetName(), DefaultInfo)
	m.Store(DefaultDebug.GetName(), DefaultDebug)
	m.Store(DefaultWarn.GetName(), DefaultWarn)
	m.Store(DefaultErr.GetName(), DefaultErr)

	m.Store("red", DefaultRed)
	m.Store("yellow", DefaultYellow)
	m.Store("blue", DefaultBlue)
	m.Store("green", DefaultGreen)
	m.Store("cyan", DefaultCyan)
	m.Store("magenta", DefaultMagenta)
}

// New 新建,传入前缀
func New(name string) *Entity {
	val, has := m.Load(name)
	if has && val != nil {
		if val, ok := val.(*Entity); ok {
			return val
		}
	}
	newEntity := NewEntity(name)
	m.Store(name, newEntity)
	return newEntity
}

// SetWriter 覆盖io.Writer
func SetWriter(fn ...io.Writer) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).SetWriter(fn...)
		return true
	})
}

// AddWriter 添加io.Writer
func AddWriter(fn ...io.Writer) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).AddWriter(fn...)
		return true
	})
}

// WriteToTCPClient 全部日志写入TCP客户端,color是否传输颜色数据
func WriteToTCPClient(addr string, color ...bool) (err error) {
	var writer io.Writer
	writer, err = NewTCPClient(addr)
	if err != nil {
		return err
	}
	if len(color) > 0 && color[0] {
		writer = NewWriteColor(writer)
	}
	AddWriter(writer)
	return nil
}

// WriteToTCPServer 全部日志写入TCP服务端,color 是否传输颜色数据
func WriteToTCPServer(port int, color ...bool) (err error) {
	var writer io.Writer
	writer, err = NewTCPServer(port)
	if err != nil {
		return err
	}
	if len(color) > 0 && color[0] {
		writer = NewWriteColor(writer)
	}
	AddWriter(writer)
	return nil
}

// WriteToHTTPServer 全部日志写入到HTTP服务端,color 是否传输颜色数据
func WriteToHTTPServer(method, url string, color ...bool) (err error) {
	var writer io.Writer
	writer, err = NewHTTPClient(method, url)
	if err != nil {
		return err
	}
	if len(color) > 0 && color[0] {
		writer = NewWriteColor(writer)
	}
	AddWriter(writer)
	return nil
}

// SetCaller 日志位置层级
func SetCaller(n int) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).SetCaller(n)
		return true
	})
}

// SetShowColor 显示颜色
func SetShowColor(b ...bool) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).SetShowColor(b...)
		return true
	})
}

// SetLevel 设置日志等级
func SetLevel(level Level) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).SetLevel(level)
		return true
	})
}

// SetLevelWithAll 设置日志等级为全部
func SetLevelWithAll() {
	SetLevel(LevelAll)
}

// SetFormatter 设置输出格式
func SetFormatter(f IFormatter) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).SetFormatter(f)
		return true
	})
}

// SetFormatterWithDefault 设置输出格式为默认
func SetFormatterWithDefault() {
	SetFormatter(FDefault)
}

func SetFormatterWithTime() {
	SetFormatter(FTime)
}

// SetSaveTime 设置保存时间,默认按天(即设置秒,用默认格式相当于1天)
func SetSaveTime(saveTime time.Duration) {
	DefaultRemoveFile.SetSaveTime(saveTime)
}

//=================================

// PrintErr 打印错误,有错误才打印
func PrintErr(err error) bool {
	if err != nil {
		DefaultErr.Println(err.Error())
	}
	return err != nil
}

// PanicErr 有错误的时候panic
func PanicErr(err error) bool {
	if err != nil {
		DefaultErr.Println(err.Error())
		panic(err)
	}
	return err != nil
}

// Spend 记录耗时,使用方式 defer Spend()()
func Spend(prefix ...interface{}) func() {
	now := time.Now()
	return func() {
		DefaultDebug.Println(fmt.Sprint(prefix...) + time.Now().Sub(now).String())
	}
}

// Trace 预设追溯 绿色
// [追溯] 2022/01/08 10:44:02 init_test.go:10:
func Trace(s ...interface{}) (int, error) {
	return DefaultTrace.Println(s...)
}

// Tracef 预设追溯 绿色
// [追溯] 2022/01/08 10:44:02 init_test.go:10:
func Tracef(format string, s ...interface{}) (int, error) {
	return DefaultTrace.Printf(format, s...)
}

// Debug 预设调试 黄色
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debug(s ...interface{}) (int, error) {
	return DefaultDebug.Println(s...)
}

// Debugf 预设调试 黄色
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debugf(format string, s ...interface{}) (int, error) {
	return DefaultDebug.Printf(format, s...)
}

// Read 预设读取 蓝色
// [读取] 2022/01/08 10:44:02 init_test.go:10:
func Read(s ...interface{}) (int, error) {
	return DefaultRead.Println(s...)
}

// Readf 预设读取 蓝色
// [读取] 2022/01/08 10:44:02 init_test.go:10:
func Readf(format string, s ...interface{}) (int, error) {
	return DefaultRead.Printf(format, s...)
}

// Write 预设写入 蓝色
// [写入] 2022/01/08 10:44:02 init_test.go:10:
func Write(s ...interface{}) (int, error) {
	return DefaultWrite.Println(s...)
}

// Writef 预设写入 蓝色
// [写入] 2022/01/08 10:44:02 init_test.go:10:
func Writef(format string, s ...interface{}) (int, error) {
	return DefaultWrite.Printf(format, s...)
}

// Info 预设信息 青色
// [信息] 2022/01/08 10:44:02 init_test.go:10:
func Info(s ...interface{}) (int, error) {
	return DefaultInfo.Println(s...)
}

// Infof 预设信息 青色
// [信息] 2022/01/08 10:44:02 init_test.go:10:
func Infof(format string, s ...interface{}) (int, error) {
	return DefaultInfo.Printf(format, s...)
}

// Warn 预设警告
// [警告] 2022/01/08 10:44:02 init_test.go:10:
func Warn(s ...interface{}) (int, error) {
	return DefaultWarn.Println(s...)
}

// Warnf 警告
func Warnf(format string, s ...interface{}) (int, error) {
	return DefaultWarn.Printf(format, s...)
}

// Err 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Err(s ...interface{}) (int, error) {
	return DefaultErr.Println(s...)
}

// Error 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Error(s ...interface{}) (int, error) {
	return DefaultErr.Println(s...)
}

// Errorf 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Errorf(format string, s ...interface{}) (int, error) {
	return DefaultErr.Printf(format, s...)
}

// Errf 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Errf(format string, s ...interface{}) (int, error) {
	return DefaultErr.Printf(format, s...)
}

// Panic 预设错误 红色
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Panic(s ...interface{}) (int, error) {
	msg := fmt.Sprint(s...)
	n, err := DefaultErr.Println(msg)
	panic(msg)
	return n, err
}

// Panicf 预设错误 红色
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Panicf(format string, s ...interface{}) (int, error) {
	msg := fmt.Sprintf(format, s...)
	n, err := DefaultErr.Println(msg)
	panic(msg)
	return n, err
}

// Fatal 预设错误 红色
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Fatal(s ...interface{}) (int, error) {
	defer os.Exit(-127)
	return DefaultErr.Println(s...)
}

// Fatalf 预设错误 红色
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Fatalf(format string, s ...interface{}) (int, error) {
	defer os.Exit(-127)
	return DefaultErr.Printf(format, s...)
}

func Red(s ...interface{}) (int, error) {
	return DefaultRed.Println(s...)
}

func Redf(format string, s ...interface{}) (int, error) {
	return DefaultRed.Printf(format, s...)
}

func Yellow(s ...interface{}) (int, error) {
	return DefaultYellow.Println(s...)
}

func Yellowf(format string, s ...interface{}) (int, error) {
	return DefaultYellow.Printf(format, s...)
}

func Blue(s ...interface{}) (int, error) {
	return DefaultBlue.Println(s...)
}

func Bluef(format string, s ...interface{}) (int, error) {
	return DefaultBlue.Printf(format, s...)
}

func Green(s ...interface{}) (int, error) {
	return DefaultGreen.Println(s...)
}

func Greenf(format string, s ...interface{}) (int, error) {
	return DefaultGreen.Printf(format, s...)
}

func Cyan(s ...interface{}) (int, error) {
	return DefaultCyan.Println(s...)
}

func Cyanf(format string, s ...interface{}) (int, error) {
	return DefaultCyan.Printf(format, s...)
}

func Magenta(s ...interface{}) (int, error) {
	return DefaultMagenta.Println(s...)
}

func Magentaf(format string, s ...interface{}) (int, error) {
	return DefaultMagenta.Printf(format, s...)
}

func Println(s ...interface{}) (int, error) {
	return fmt.Println(s...)
}

func Printf(format string, s ...interface{}) (int, error) {
	return fmt.Printf(format, s...)
}
