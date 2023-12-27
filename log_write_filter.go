package logs

import (
	"io"
	"regexp"
)

//==============================WriteFilter==============================

func NewFilter(w io.Writer) *Filter {
	return &Filter{Writer: w}
}

type Filter struct {
	io.Writer
	reg    *regexp.Regexp //正则表达式
	enable bool           //是否启用
}

func (this *Filter) Enable(b ...bool) {
	this.enable = len(b) == 0 || b[0]
}

// SetRegexp 设置正则表达式
func (this *Filter) SetRegexp(reg string) {
	this.reg, _ = regexp.Compile(reg)
}

// SetLike 设置模糊搜索
func (this *Filter) SetLike(like string) {
	this.SetRegexp(`.*` + like + `.*`)
}

func (this *Filter) Valid(p []byte) bool {
	return this.reg == nil || this.reg.Match(p)
}

func (this *Filter) Write(p []byte) (int, error) {
	if this.Valid(p) {
		return this.Writer.Write(p)
	}
	return len(p), nil
}
