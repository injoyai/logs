package logs

import (
	"bytes"
	"log"
)

/*
	只做输出格式化,不涉及颜色
	默认输出:
		[信息] 2006-01-02 15:04:05 logs_format:12 [tag] 默认输出


*/

// IFormatter 定义输出接口
type IFormatter interface {
	Formatter(e *Entity, msg string) string
}

var (
	// DefaultFormatter 默认格式化可修改
	DefaultFormatter = &formatter{flag: log.Ldate | log.Ltime | log.Lshortfile}

	TimeFormatter FormatFunc = timeFormatter
)

// 默认输出
type formatter struct {
	flag      int //当formatter为nil时使用
	formatter func(e *Entity, msg string) string
}

// SetFlag 设置flag
func (this *formatter) SetFlag(flag int) *formatter {
	this.flag = flag
	return this
}

// SetFormatter 设置数据格式函数
func (this *formatter) SetFormatter(formatter func(e *Entity, msg string) string) *formatter {
	this.formatter = formatter
	return this
}

// Formatter 默认输出格式函数,实现接口
func (this *formatter) Formatter(e *Entity, msg string) string {
	if this.formatter != nil {
		return this.formatter(e, msg)
	}
	writer := bytes.NewBuffer(nil)
	var tag string
	for i, v := range e.Tag {
		tag += "[" + v + "]"
		if i == len(e.Tag)-1 {
			tag += " "
		}
	}
	msg = tag + msg
	prefix := ""
	if len(e.Name) > 0 {
		prefix = "[" + e.Name + "]"
	}

	hasLn := len(msg) > 0 && msg[len(msg)-1] == '\n'
	_ = log.New(writer, prefix, this.flag).Output(e.GetCaller(), msg)
	msg = writer.String()
	if len(msg) > 0 && msg[len(msg)-1] == '\n' && !hasLn {
		//去除最后一个换行,,如果没有\n,log会自动添加一个\n
		msg = msg[:len(msg)-1]
	}
	return msg
}

type FormatFunc func(e *Entity, msg string) string

func (thiS FormatFunc) Formatter(e *Entity, msg string) string {
	return thiS(e, msg)
}

func timeFormatter(e *Entity, msg string) string {
	writer := bytes.NewBuffer(nil)
	var tag string
	for i, v := range e.Tag {
		tag += "[" + v + "]"
		if i == len(e.Tag)-1 {
			tag += " "
		}
	}
	msg = tag + msg
	if len(e.Name) > 0 {
		msg = "[" + e.Name + "] " + msg
	}

	hasLn := len(msg) > 0 && msg[len(msg)-1] == '\n'
	_ = log.New(writer, "", log.Ltime).Output(e.GetCaller(), msg)
	msg = writer.String()
	if len(msg) > 0 && msg[len(msg)-1] == '\n' && !hasLn {
		msg = msg[:len(msg)-1]
	}
	return msg
}
