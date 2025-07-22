package logs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

//==============================WriteTrunk==============================

// NewTrunk 消息总线,发布和订阅
func newTrunk(cap int) *trunk {
	return &trunk{
		c: make(chan []byte, cap),
	}
}

// Trunk 消息总线,发布和订阅
type trunk struct {
	c         chan []byte
	subscribe []*trunkSubscribe
	sync.Mutex
}

// Write 实现io.Writer接口
func (this *trunk) Write(p []byte) (int, error) {
	this.Publish(p)
	return len(p), nil
}

// Publish 发布接口输入
func (this *trunk) Publish(data ...[]byte) {
	for _, sub := range this.subscribe {
		if sub != nil {
			sub.try(data...)
		}
	}
}

// Subscribe 订阅消息总线
func (this *trunk) Subscribe(bufSize int, handler func(data []byte)) string {
	key := fmt.Sprintf("%p-%p-%d", this, handler, time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	sub := &trunkSubscribe{
		key:     key,
		handler: handler,
		c:       make(chan []byte, bufSize),
		ctx:     ctx,
		cancel:  cancel,
	}
	go sub.run()
	this.subscribe = append(this.subscribe, sub)
	return key
}

// Unsubscribe 取消订阅
func (this *trunk) Unsubscribe(key string) bool {
	if len(key) == 0 {
		return false
	}
	this.Lock()
	defer this.Unlock()
	for i, v := range this.subscribe {
		if v.key == key {
			this.subscribe = append(this.subscribe[:i], this.subscribe[i+1:]...)
			v.cancel()
			return true
		}
	}
	return false
}

type trunkSubscribe struct {
	key     string
	handler func(data []byte)
	c       chan []byte
	ctx     context.Context
	cancel  context.CancelFunc
}

func (this *trunkSubscribe) try(data ...[]byte) {
	for _, v := range data {
		select {
		case <-this.ctx.Done():
			return
		case this.c <- v:
		default:
		}
	}
}

func (this *trunkSubscribe) run() {
	for {
		select {
		case <-this.ctx.Done():
			return
		case data := <-this.c:
			if this.handler != nil {
				this.handler(data)
			}
		}
	}
}
