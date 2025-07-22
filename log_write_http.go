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

func NewHTTPClient(method, url string) io.Writer {
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
		method: method,
		url:    url,
		Chan:   newChan(context.Background(), 100),
	}
	w.Chan.handler = func(ctx context.Context, count int, bs []byte) {
		req, err := http.NewRequest(w.method, w.url, bytes.NewBuffer(bs))
		if err == nil {
			w.Client.Do(req)
		}
	}
	return w
}

type httpClient struct {
	*http.Client
	method string
	url    string
	*Chan
}
