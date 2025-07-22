package logs

import (
	"context"
)

func newChan(ctx context.Context, cap int) *Chan {
	data := &Chan{
		c:   make(chan []byte, cap),
		ctx: ctx,
	}
	go data.run(ctx)
	return data
}

type Chan struct {
	c       chan []byte                                     //通道
	handler func(ctx context.Context, count int, bs []byte) //数据处理
	ctx     context.Context
}

func (this *Chan) Write(p []byte) (int, error) {
	return len(p), this.Try(p)
}

// Try 尝试加入队列(如果满了则忽略)
func (this *Chan) Try(data ...[]byte) error {
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

func (this *Chan) run(ctx context.Context) {
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
