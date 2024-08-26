package fasthttp_client

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func (c *FasthttpClient) SetUrlRequest(url string, req *fasthttp.Request) error {
	var url_clear []byte
	for i := 0; i < len([]byte(url)); i++ {
		b := url[i]
		if b < ' ' || b == 0x7f {
			continue
		}
		url_clear = append(url_clear, b)
	}

	req.SetRequestURIBytes(url_clear)

	if len(req.URI().Host()) > 0 {
		return nil
	}

	return fmt.Errorf("Failed set URI (%v)", url)
}

// ====================== Request and Response ======================

func (c *FasthttpClient) GetNewRequest() *fasthttp.Request {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return fasthttp.AcquireRequest()
}

func (c *FasthttpClient) GetNewResponse() *fasthttp.Response {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return fasthttp.AcquireResponse()
}

func (c *FasthttpClient) GetNewReqAndResp() (*fasthttp.Request, *fasthttp.Response) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
}

func (c *FasthttpClient) RealeseRequest(req *fasthttp.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	fasthttp.ReleaseRequest(req)
}

func (c *FasthttpClient) RealeseResponse(resp *fasthttp.Response) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	fasthttp.ReleaseResponse(resp)
}

func (c *FasthttpClient) RealeseReqAndRes(req *fasthttp.Request, resp *fasthttp.Response) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}

func (c *FasthttpClient) ResetRequest(req *fasthttp.Request) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	req.Reset()
}

func (c *FasthttpClient) ResetResponse(resp *fasthttp.Response) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	resp.Reset()
}

func (c *FasthttpClient) ResetReqResp(req *fasthttp.Request, resp *fasthttp.Response) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	req.Reset()
	resp.Reset()
}
