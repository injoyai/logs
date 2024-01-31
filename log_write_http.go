package logs

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

//==============================WriteHTTP==============================

type httpClient struct {
	*http.Client
	method string
	url    string
	ch     *_chan

	//消息序号,用于监测数据是否丢失
	//数据是放入一定大小的缓存中,如果缓存满了,则会丢弃数据,防止阻塞
	//如果接收端发现消息序号断层了,说明数据有丢弃情况,
	//一般不会出现,为了以防万一,所以增加这个字段,加个数据的最前方1字节
	index    uint8
	useIndex bool
}

func (this *httpClient) Write(p []byte) (int, error) {
	if this.useIndex {
		p = append([]byte{this.index}, p...)
		this.index++
	}
	return len(p), this.ch.Try(p)
}

func NewHTTPClient(method, url string, useIndex ...bool) (io.Writer, error) {
	w := &httpClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: time.Second * 10,
		},
		method:   method,
		url:      url,
		ch:       newChan(context.Background(), 100),
		useIndex: len(useIndex) > 0 && useIndex[0],
	}
	w.ch.handler = func(ctx context.Context, count int, data interface{}) {
		bs := data.([]byte)
		req, err := http.NewRequest(w.method, w.url, bytes.NewBuffer(bs))
		if err == nil {
			w.Client.Do(req)
		}
	}
	return w, nil
}
