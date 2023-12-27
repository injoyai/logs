/*****************************************************************************
*名称:	日志包
*功能:	日志的打印,写入文件
*作者:	钱纯净
******************************************************************************/

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
	m            = sync.Map{}
	DefaultTrace = NewEntity("追溯").setCaller(1).SetColor(color.FgGreen)
	DefaultWrite = NewEntity("写入").setCaller(1).SetColor(color.FgBlue)
	DefaultRead  = NewEntity("读取").setCaller(1).SetColor(color.FgBlue)
	DefaultInfo  = NewEntity("信息").setCaller(1).SetColor(color.FgCyan)
	DefaultDebug = NewEntity("调试").setCaller(1).SetColor(color.FgYellow)
	DefaultWarn  = NewEntity("警告").setCaller(1).SetColor(color.FgMagenta)
	DefaultErr   = NewEntity("错误").setCaller(1).SetColor(color.FgRed).WriteToFile(DefaultDir, DefaultLayout)

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
	SetFormatter(DefaultFormatter)
}

// SetSaveTime 设置保存时间,默认按天(即设置秒,用默认格式相当于1天)
func SetSaveTime(saveTime time.Duration) {
	DefaultRemoveFile.SetSaveTime(saveTime)
}

//=================================

// PrintErr 打印错误,有错误才打印
func PrintErr(err error) bool {
	if err != nil {
		DefaultErr.Write(LevelError, err.Error())
	}
	return err != nil
}

// PanicErr 有错误的时候panic
func PanicErr(err error) bool {
	if err != nil {
		DefaultErr.Write(LevelError, err.Error())
		panic(err)
	}
	return err != nil
}

// Trace 预设追溯 蓝色
// [追溯] 2022/01/08 10:44:02 init_test.go:10:
func Trace(s ...interface{}) {
	DefaultTrace.Write(LevelTrace, s...)
}

// Tracef 预设细微 蓝色
// [追溯] 2022/01/08 10:44:02 init_test.go:10:
func Tracef(format string, s ...interface{}) {
	DefaultTrace.Writef(LevelTrace, format, s...)
}

// Read 预设读取 蓝色
// [读取] 2022/01/08 10:44:02 init_test.go:10:
func Read(s ...interface{}) {
	DefaultRead.Write(LevelRead, s...)
}

// Readf 预设读取 蓝色
// [读取] 2022/01/08 10:44:02 init_test.go:10:
func Readf(format string, s ...interface{}) {
	DefaultRead.Writef(LevelRead, format, s...)
}

// Write 预设写入 蓝色
// [写入] 2022/01/08 10:44:02 init_test.go:10:
func Write(s ...interface{}) {
	DefaultWrite.Write(LevelWrite, s...)
}

// Writef 预设写入 蓝色
// [写入] 2022/01/08 10:44:02 init_test.go:10:
func Writef(format string, s ...interface{}) {
	DefaultWrite.Writef(LevelWrite, format, s...)
}

// Info 预设信息 青色
// [信息] 2022/01/08 10:44:02 init_test.go:10:
func Info(s ...interface{}) {
	DefaultInfo.Write(LevelInfo, s...)
}

// Infof 预设信息 青色
// [信息] 2022/01/08 10:44:02 init_test.go:10:
func Infof(format string, s ...interface{}) {
	DefaultInfo.Writef(LevelInfo, format, s...)
}

// Debug 预设调试 黄色
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debug(s ...interface{}) {
	DefaultDebug.Write(LevelDebug, s...)
}

// Debugf 预设调试 黄色
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debugf(format string, s ...interface{}) {
	DefaultDebug.Writef(LevelDebug, format, s...)
}

// Warn 预设警告
// [警告] 2022/01/08 10:44:02 init_test.go:10:
func Warn(s ...interface{}) {
	DefaultWarn.Write(LevelWarn, s...)
}

// Warnf 警告
func Warnf(format string, s ...interface{}) {
	DefaultWarn.Writef(LevelWarn, format, s...)
}

// Err 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Err(s ...interface{}) {
	DefaultErr.Write(LevelError, s...)
}

// Error 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Error(s ...interface{}) {
	DefaultErr.Write(LevelError, s...)
}

// Errorf 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Errorf(format string, s ...interface{}) {
	DefaultErr.Writef(LevelError, format, s...)
}

// Errf 预设错误 红色 写入文件
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Errf(format string, s ...interface{}) {
	DefaultErr.Writef(LevelError, format, s...)
}

func Spend(prefix ...interface{}) func() {
	now := time.Now()
	return func() {
		DefaultDebug.Write(LevelDebug, fmt.Sprint(prefix...), time.Now().Sub(now))
	}
}

// Panic 预设调试 红色
// [错误] 2022/01/08 10:44:02 init_test.go:10:
func Panic(s ...interface{}) {
	msg := fmt.Sprint(s...)
	DefaultErr.Write(LevelError, msg)
	panic(msg)
}

// Panicf 预设调试 红色
// [致命] 2022/01/08 10:44:02 init_test.go:10:
func Panicf(format string, s ...interface{}) {
	msg := fmt.Sprintf(format, s...)
	DefaultErr.Write(LevelError, msg)
	panic(msg)
}

// Fatal 预设调试 红色
// [致命] 2022/01/08 10:44:02 init_test.go:10:
func Fatal(s ...interface{}) {
	DefaultErr.Write(LevelError, s...)
	os.Exit(-127)
}

// Fatalf 预设调试 红色
// [致命] 2022/01/08 10:44:02 init_test.go:10:
func Fatalf(format string, s ...interface{}) {
	DefaultErr.Writef(LevelError, format, s...)
	os.Exit(-127)
}
