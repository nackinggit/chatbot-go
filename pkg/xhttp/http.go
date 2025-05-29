package xhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
)

type Client struct {
	httpclient *http.Client
}

func NewClient(hc *http.Client) *Client {
	return &Client{
		httpclient: hc,
	}
}

var innerclient = &Client{
	httpclient: &http.Client{},
}

func (c *Client) DoAndBind(request *http.Request, target any) error {
	bs, err := c.do(request)
	xlog.Debugf("[xhttp] request: %v, resp: %v, err: %v", request.URL, string(bs), err)
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

func (c *Client) PostJsonAndBind(request *JsonPostRequest, target any) error {
	r, err := request.toRequest()
	if err != nil {
		return err
	}
	return c.DoAndBind(r, target)
}

func (c *Client) GetAndBind(ctx context.Context, url string, headers http.Header, target any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	if headers != nil {
		req.Header = headers
	}
	return c.DoAndBind(req, target)
}

func PostJsonAndBind[T any](request *JsonPostRequest, target *T) error {
	xlog.Debugf("Post request: %v", util.JsonString(request))
	r, err := request.toRequest()
	if err != nil {
		return err
	}
	return innerclient.DoAndBind(r, target)
}

func GetAndBind[T any](ctx context.Context, url string, headers http.Header, target *T) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	if headers != nil {
		req.Header = headers
	}
	return innerclient.DoAndBind(req, target)
}

type JsonPostRequest struct {
	Headers http.Header
	Body    any
	Url     string
	Context context.Context
}

func (r *JsonPostRequest) toRequest() (*http.Request, error) {
	ctx := r.Context
	if ctx == nil {
		ctx = context.Background()
	}
	bs, err := util.Marshal(r.Body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Url, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}
	if r.Headers != nil {
		request.Header = r.Headers
	}
	return request, nil
}
