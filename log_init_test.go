package logs

import (
	"testing"
	"time"
)

func TestSetWriter(t *testing.T) {
	f := NewFile("./output/logs/debug.log", 2<<20)
	DefaultDebug.SetWriter(f)
	for {
		Debug(make([]byte, 1024))
		<-time.After(time.Millisecond * 10)
	}
}
