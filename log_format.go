package logs

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"
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
	DefaultFormatter = FDefault

	// FDefault 默认格式化
	FDefault = &formatter{flag: log.Ldate | log.Ltime | log.Lshortfile}

	// TimeFormatter 时间格式化
	TimeFormatter = FTime

	// FTime 时间格式化
	FTime FormatFunc = timeFormatter

	// FJson json格式化
	FJson FormatFunc = jsonFormatter
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
	msg = buildTag(e.Tag) + msg
	prefix := ""
	if len(e.Name) > 0 {
		prefix = "[" + e.Name + "] "
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

func (this FormatFunc) Formatter(e *Entity, msg string) string {
	return this(e, msg)
}

func jsonFormatter(e *Entity, msg string) string {
	logMap := map[string]interface{}{
		"level": e.Name,
		"time":  time.Now().Format(time.RFC3339),
		"tag":   e.Tag,
		"msg":   msg,
	}
	b, _ := json.Marshal(logMap)
	return string(b)
}

func timeFormatter(e *Entity, msg string) string {
	writer := bytes.NewBuffer(nil)
	msg = buildTag(e.Tag) + msg
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

func buildTag(tags []string) string {
	var b strings.Builder
	for _, t := range tags {
		b.WriteString("[" + t + "]")
	}
	return b.String()
}
