package fasthttp_client

import (
	"fmt"
	"github.com/GoHippo/network/proxy/checker_proxy"
	"github.com/GoHippo/network/proxy/config"
	"github.com/GoHippo/slogpretty/sl"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type FasthttpClient struct {
	*fasthttp.Client
	mutex       *sync.Mutex
	ProxyConfig config.ProxyConfig
	FastHttpClientOptions
}

func (c *FasthttpClient) Do(req *fasthttp.Request, resp *fasthttp.Response, option DoOption) (body_decode string, err error) {
	var op = `network.fasthttp.Do`

	if len(req.URI().Host()) == 0 {
		return "", fmt.Errorf("[Network][%v] Host is nil", option.ID)
	}

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

		if option.ErrCounter != nil {
			option.ErrCounter.AddCountNetworkErr(err)
		}

		if err == fasthttp.ErrTimeout {

			if option.DoCountReconnecting < 1 {
				return fmt.Errorf("[%v] %v", option.ID, err.Error())
			}

			option.DoCountReconnecting -= 1

			if c.ProxyUse {
				c.Log.Error(fmt.Sprintf("[Network][%v] Timeout к серверу истек, делаю переподключение c новым proxy!", option.ID))
				return c.do_with_new_proxy(req, resp, option)
			}

			c.Log.Error(fmt.Sprintf("[Network][%v] Timeout к серверу истек, делаю переподключение!", option.ID))
			return c.do(req, resp, option)
		}

		if c.checkErrConn(err) {

			if c.ProxyUse {
				c.ProxyService.DeleteProxy(c.ProxyConfig)
				return c.do_with_new_proxy(req, resp, option)
			}

			return fmt.Errorf("[%v] %v", option.ID, err.Error())
		}

		if option.DoCountReconnecting > 0 {
			c.Log.Error(fmt.Sprintf("[Network][%v] Попытка переподключения. Ошибка при запросе: err-%v host-%v path:%v", option.ID, err.Error(), string(req.URI().Host()), string(req.URI().Path())))
			option.DoCountReconnecting -= 1

			return c.do(req, resp, option)
		}

		return fmt.Errorf("[%v] %v", option.ID, err.Error())
	}
	// resp.LocalAddr()
	return nil
}

func (c *FasthttpClient) do_with_new_proxy(req *fasthttp.Request, resp *fasthttp.Response, option DoOption) error {
	// pc := c.ProxyConfig
	// defer c.proxyService.FreeProxy(pc)

	c.ProxyService.FreeProxy(c.ProxyConfig)

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
			c.Log.Error("[CheckErrInternet][%v]%v", err.Error())

			if c.ProxyUse {
				c.ProxyService.DeleteProxy(c.ProxyConfig)
			}

			return true
		}
	}

	if c.ProxyUse {
		return checker_proxy.CheckTreeDomains(c.Client, c.FastHttpClientOptions.DialTimeout) != nil
	}
	return checker_proxy.CheckTreeDomains(c.Client, time.Second*5) != nil
}
