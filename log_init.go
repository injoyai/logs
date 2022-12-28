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
	"sync"
	"time"
)

const (
	DefaultDir    = "./output/logs/"
	DefaultLayout = "2006-01-02/{name}_15.log"
)

var (
	m            = sync.Map{}
	DefaultTest  = NewEntity("测试").setCaller(1).SetColor(color.FgBlue)
	DefaultInfo  = NewEntity("信息").setCaller(1).SetColor(color.FgCyan)
	DefaultDebug = NewEntity("调试").setCaller(1).SetColor(color.FgYellow)
	DefaultWarn  = NewEntity("警告").setCaller(1).SetColor(color.FgMagenta)
	DefaultErr   = NewEntity("错误").setCaller(1).SetColor(color.FgRed).WriteToFile(DefaultDir, DefaultLayout)

	// Trunk 消息总线,公共Writer,
	Trunk = newTrunk(1000)

	// Hook 计划支持控制台筛选,tcp筛选...
	Hook = newHook()

	// Stdout 系统标准输出, NewWriteColor 支持颜色输出
	Stdout = newStdout()
)

func init() {
	color.NoColor = false
	m.Store(DefaultTest.GetName(), DefaultTest)
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

// SetFormatter 设置输出格式
func SetFormatter(f IFormatter) {
	m.Range(func(key, value interface{}) bool {
		value.(*Entity).SetFormatter(f)
		return true
	})
}

// SetSaveTime 设置保存时间,默认按天(即设置秒,用默认格式相当于1天)
func SetSaveTime(saveTime time.Duration) {
	DefaultRemoveFile.SetSaveTime(saveTime)
}

//=================================

// PrintErr 打印错误,有错误才打印
func PrintErr(err error) bool {
	if err != nil {
		DefaultErr.Write(err.Error())
	}
	return err != nil
}

// Test 预设测试 蓝色
//[测试] 2022/01/08 10:44:02 init_test.go:10:
func Test(s ...interface{}) {
	DefaultTest.Write(s...)
}

// Testf 预设测试 蓝色
//[测试] 2022/01/08 10:44:02 init_test.go:10:
func Testf(format string, s ...interface{}) {
	DefaultTest.Write(fmt.Sprintf(format, s...))
}

// Warn 预设警告
//[警告] 2022/01/08 10:44:02 init_test.go:10:
func Warn(s ...interface{}) {
	DefaultWarn.Write(s...)
}

// Warnf 警告
func Warnf(format string, s ...interface{}) {
	DefaultWarn.Write(fmt.Sprintf(format, s...))
}

// Err 预设错误 红色 写入文件
//[错误] 2022/01/08 10:44:02 init_test.go:10:
func Err(s ...interface{}) {
	DefaultErr.Write(s...)
}

// Error 预设错误 红色 写入文件
//[错误] 2022/01/08 10:44:02 init_test.go:10:
func Error(s ...interface{}) {
	DefaultErr.Write(s...)
}

// Errorf 预设错误 红色 写入文件
//[错误] 2022/01/08 10:44:02 init_test.go:10:
func Errorf(format string, s ...interface{}) {
	DefaultErr.Write(fmt.Sprintf(format, s...))
}

// Errf 预设错误 红色 写入文件
//[错误] 2022/01/08 10:44:02 init_test.go:10:
func Errf(format string, s ...interface{}) {
	DefaultErr.Write(fmt.Sprintf(format, s...))
}

// Info 预设信息 青色
//[信息] 2022/01/08 10:44:02 init_test.go:10:
func Info(s ...interface{}) {
	DefaultInfo.Write(s...)
}

// Infof 预设信息 青色
//[信息] 2022/01/08 10:44:02 init_test.go:10:
func Infof(format string, s ...interface{}) {
	DefaultInfo.Write(fmt.Sprintf(format, s...))
}

// Debug 预设调试 黄色
//[调试] 2022/01/08 10:44:02 init_test.go:10:
func Debug(s ...interface{}) {
	DefaultDebug.Write(s...)
}

// Debugf 预设调试 黄色
// [调试] 2022/01/08 10:44:02 init_test.go:10:
func Debugf(format string, s ...interface{}) {
	DefaultDebug.Write(fmt.Sprintf(format, s...))
}
