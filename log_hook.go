package logs

import (
	"regexp"
	"time"
)

type HookMessage struct {
	Entity  *Entity
	Message string
	Time    time.Time
}

type hook struct {
	reg *regexp.Regexp //正则表达式
}

// Hook 名字待定 符合 条件的数据
func (this *hook) Hook(p []byte) bool {
	return this.reg == nil || this.reg.Match(p)
}

func (this *hook) Regexp(s string) error {
	reg, err := regexp.Compile(s)
	if err == nil {
		this.reg = reg
	}
	return err
}

func newHook() *hook {
	return &hook{}
}
