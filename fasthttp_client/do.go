package fasthttp_client

import (
	"fmt"
	"github.com/GoHippo/network/proxy/checker_proxy"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"github.com/GoHippo/slogpretty/sl"
	"strings"
	"sync"
	"time"
	
	"github.com/valyala/fasthttp"
)

type FasthttpClient struct {
	*fasthttp.Client
	mutex       *sync.Mutex
	proxyConfig proxy_jar.ProxyConfig
	FastHttpClientOptions
}

func (c *FasthttpClient) Do(req *fasthttp.Request, resp *fasthttp.Response, option DoOption) (body_decode string, err error) {
	var op = `network.fasthttp.Do`
	
	defer req.Reset()
	
	if option.Jar != nil {
		option.Jar.SetCookiesReq(req)
		defer option.Jar.AddFromResponse(req.URI(), resp)
	}
	
	if option.DoCountReconnecting == 0 {
		option.DoCountReconnecting = c.FastHttpClientOptions.CountReconnections
	}
	
	err = c.do(req, resp, option)
	if err != nil {
		c.Log.Error(op, sl.Err(err))
		return "", err
	}
	
	if !option.NotBodyDecode {
		body, err := resp.BodyUncompressed()
		if err != nil {
			err = fmt.Errorf("[compress][%v] Ошибка при распаковке в поиске запроса сообщения err:%v content:%v", option.ID, err, resp.Header.ContentEncoding())
			c.Log.Error(err.Error())
			return string(resp.Body()), nil
		}
		
		// body = bytes.TrimPrefix(body, []byte(`)]}'`))
		
		body_decode = string(body)
	}
	
	return body_decode, nil
}

func (c *FasthttpClient) do(req *fasthttp.Request, resp *fasthttp.Response, option DoOption) error {
	
	err := c.Client.DoTimeout(req, resp, c.FastHttpClientOptions.DialTimeout)
	if err != nil {
		
		if err == fasthttp.ErrNoFreeConns {
			time.Sleep(time.Millisecond * 20)
			return c.do(req, resp, option)
		}
		
		if err == fasthttp.ErrTimeout {
			
			if option.ErrCounter != nil {
				option.ErrCounter.AddCountNetworkErr(err)
			}
			
			c.Log.Error(fmt.Sprintf("[Network][%v] Timeout к серверу истек, делаю переподключение c новым proxy!", option.ID))
			option.DoCountReconnecting -= 1
			
			if c.ProxyUse {
				return c.do_with_new_proxy(req, resp, option)
			}
			
			return c.do(req, resp, option)
		}
		
		if c.checkErrConn(err) {
			
			if option.ErrCounter != nil {
				option.ErrCounter.AddCountNetworkErr(err)
			}
			
			if c.ProxyUse {
				cli, errCli := NewFasthttpClient(c.FastHttpClientOptions)
				if errCli != nil {
					return fmt.Errorf("[%v] Ошибка Do. %v: Ошибка при создании нового клиента. %v", option.ID, err, errCli)
				}
				c = cli
				return c.do(req, resp, option)
			}
			return fmt.Errorf("[%v] %v", option.ID, err.Error())
		}
		
		if option.DoCountReconnecting != 0 {
			if option.ErrCounter != nil {
				option.ErrCounter.AddCountNetworkErr(err)
			}
			
			c.Log.Error(fmt.Sprintf("[Network][%v] Попытка переподключения. Ошибка при запросе: err-%v host-%v path:%v", option.ID, err.Error(), string(req.URI().Host()), string(req.URI().Path())))
			option.DoCountReconnecting -= 1
			return c.do(req, resp, option)
		}
		
		if c.checkErrAll(err) {
			if option.ErrCounter != nil {
				option.ErrCounter.AddCountNetworkErr(err)
			}
			
			if c.ProxyUse {
				return c.do_with_new_proxy(req, resp, option)
			}
			return fmt.Errorf("[%v] %v", option.ID, err.Error())
		}
		
		return fmt.Errorf("[%v] %v", option.ID, err.Error())
	}
	// resp.LocalAddr()
	return nil
}

func (c *FasthttpClient) do_with_new_proxy(req *fasthttp.Request, resp *fasthttp.Response, option DoOption) error {
	// pc := c.proxyConfig
	// defer c.proxyService.FreeProxy(pc)
	
	c.ProxyService.FreeProxy(c.proxyConfig)
	option.DoCountReconnecting -= 1
	
	cli, errCli := NewFasthttpClient(c.FastHttpClientOptions)
	if errCli != nil {
		return fmt.Errorf("[%v] Ошибка при создании нового клиента. %v", option.ID, errCli)
	}
	c = cli
	return c.do(req, resp, option)
}

// ====================== Error Network ======================

var (
	ArrErrInternet = []string{
		`connectex: An attempt was made to access a socket in a way forbidden by its access permissions.`,
		`connectex: A socket operation was attempted to an unreachable host.`,
		`connectex: No connection`,
		`connectex: `,
		`An existing connection was forcibly closed by the remote host`,
		`could not connect to proxy`,
		// `no such host`,
	}
)

// CheckErrInternet ищет ошибку с соеденением. Вернет true, если обнаружит.
// Удалить BadProxy, если после проверки обнаружет ошибку
func (c *FasthttpClient) checkErrConn(err error) bool {
	for _, e := range ArrErrInternet {
		if strings.Contains(err.Error(), e) {
			c.Log.Error("[CheckErrInternet] " + err.Error())
			
			if c.ProxyUse {
				c.ProxyService.DeleteProxy(c.proxyConfig)
			}
			
			return true
		}
	}
	
	switch c.ProxyUse {
	case true:
		
		err = checker_proxy.CheckTreeDomains(c.Client, c.FastHttpClientOptions.DialTimeout)
		if err != nil {
			c.ProxyService.DeleteProxy(c.proxyConfig)
		}
		return err != nil
	
	default:
		return checker_proxy.CheckTreeDomains(c.Client, time.Second*5) != nil
	}
	
}

// ====================== Error All ======================

var (
	ArrErrAll = []string{
		`no such host`,
		`the server closed connection before returning the first response byte.`,
	}
)

// CheckErrInternet ищет ошибку с соеденением. Вернет true, если обнаружит.
// Удалить BadProxy, если после проверки обнаружет ошибку
func (c *FasthttpClient) checkErrAll(err error) bool {
	for _, e := range ArrErrAll {
		if strings.Contains(err.Error(), e) {
			c.Log.Error("[CheckErrAll] " + err.Error())
			
			return true
		}
	}
	return false
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
