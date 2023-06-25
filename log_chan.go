package logs

import (
	"context"
)

type _chan struct {
	c       chan interface{}                                       //通道
	handler func(ctx context.Context, count int, data interface{}) //数据处理
	ctx     context.Context
}

func newChan(ctx context.Context, cap int) *_chan {
	data := &_chan{
		c:   make(chan interface{}, cap),
		ctx: ctx,
	}
	go data.run(ctx)
	return data
}

// Try 尝试加入队列(如果满了则忽略)
func (this *_chan) Try(data ...interface{}) error {
	for _, v := range data {
		select {
		case <-this.ctx.Done():
			return nil
		case this.c <- v:
		default:
			//尝试加入队列失败
		}
	}
	return nil
}

// Do 添加数据,通道关闭,返回错误信息
// @data,数据任意类型
func (this *_chan) Do(data ...interface{}) error {
	for _, v := range data {
		select {
		case <-this.ctx.Done():
			return nil
		case this.c <- v:
		}
	}
	return nil
}

func (this *_chan) run(ctx context.Context) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		case v := <-this.c:
			if this.handler != nil {
				this.handler(ctx, i, v)
			}
		}
	}
}
