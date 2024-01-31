package logs

import (
	"fmt"
	"io"
	"os"
	"sync"
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
	Filter *Filter
	once   sync.Once
}

func (this *stdout) Write(p []byte) (int, error) {
	if this.Filter != nil {
		if this.Filter.Valid(p) {
			return this.Writer.Write(p)
		}
	}
	return len(p), nil
}

// Color 暂时用这个方法判断是否支持颜色
func (this *stdout) Color() bool { return true }

// EnableFilter 启用数据过滤
func (this *stdout) EnableFilter() {
	this.once.Do(func() {
		go func(s *stdout) {
			for {
				var input string
				fmt.Scanln(&input)
				s.Filter.SetLike(input)
			}
		}(this)
	})
}

// newStdout 新建带Hook的标准输出
func newStdout() *stdout {
	s := &stdout{Writer: os.Stdout, Filter: NewFilter(nil)}
	return s
}
