package xhttp

import (
	"io"
	"net/http"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
)

type Client struct {
	httpclient *http.Client
}

var innerclient = &Client{
	httpclient: &http.Client{},
}

func (c *Client) DoAndBind(request *http.Request, target any) error {
	bs, err := c.do(request)
	if err != nil {
		xlog.Warnf("http 请求失败: err: %v, response: %v", err, string(bs))
	}
	return util.Unmarshal(bs, target)
}

func (c *Client) do(req *http.Request) (bs []byte, err error) {
	resp, err := c.httpclient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
		bs, ioerr := io.ReadAll(resp.Body)
		if err != nil {
			return bs, err
		}
		return bs, ioerr
	}
	return bs, err
}

func DoAndBind[T any](request *http.Request, target *T) error {
	return innerclient.DoAndBind(request, target)
}
