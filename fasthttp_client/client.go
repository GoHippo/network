package fasthttp_client

import (
	"crypto/tls"
	"fmt"
	"github.com/valyala/fasthttp"
	"log/slog"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"sync"
	"time"
)

// NewFasthttpClient Создает клиент fasthttp
// Если в конфиге включен прокси, то вернет клиент с прокси.
// Если прокси будет пустое и активна "блокировка без прокси" в cfg, то вернет ошибку.
// Нужно закрыть клиент, чтобы высвободить Proxy для других клиентов.
// count_reconnections - количество переподключений для ошибок, что не связанны с интернетом.
type FastHttpClientOptions struct {
	DialTimeout              time.Duration
	MaxConnsPerHost          int
	Log                      *slog.Logger
	ProxyService             ProxyService
	ProxyUse                 bool
	BlockNetworkWithoutProxy bool // when null proxy in jar
	CountReconnections       int
}

type ProxyService interface {
	DeleteProxy(p proxy_jar.ProxyConfig)
	FreeProxy(p proxy_jar.ProxyConfig)
	GetCountProxy() int
	GetCountProxyImap() int
	GetFasthttpProxy(dialTimeout time.Duration) (fasthttp.DialFunc, proxy_jar.ProxyConfig, error)
}

func NewFasthttpClient(options FastHttpClientOptions) (*FasthttpClient, error) {
	
	if err := options.CheckNew(); err != nil {
		return nil, err
	}
	
	client := &FasthttpClient{
		Client: &fasthttp.Client{
			// max размер буффера для пакета запроса
			ReadBufferSize:  15 * 1024,
			MaxConnsPerHost: options.MaxConnsPerHost,
			// MaxConnWaitTimeout:  5 * time.Second,
			// MaxIdleConnDuration: 5 * time.Second,
			MaxConnWaitTimeout: 200 * time.Second,
			TLSConfig:          &tls.Config{InsecureSkipVerify: true},
		},
		mutex:                 &sync.Mutex{},
		FastHttpClientOptions: options,
	}
	
	if client.ProxyUse && client.ProxyService.GetCountProxy() != 0 {
		dial, proxyConfig, err := options.ProxyService.GetFasthttpProxy(options.DialTimeout)
		if err != nil {
			if options.BlockNetworkWithoutProxy {
				return nil, fmt.Errorf("Ошибка. Прокси закончились. Стоит запрет на использование сети без прокси.")
			}
			return client, nil
		}
		client.Dial = dial
		client.proxyConfig = proxyConfig
		client.ProxyUse = true
	}
	
	return client, nil
}

func (fc *FasthttpClient) Close() {
	if fc.ProxyUse {
		fc.ProxyService.FreeProxy(fc.proxyConfig)
		fc.ProxyUse = false
	}
	fc.Client.CloseIdleConnections()
}

func (fco *FastHttpClientOptions) CheckNew() error {
	var op = "fasthttp_client/client/NewFasthttpClient"
	
	if fco.DialTimeout == 0 {
		fco.DialTimeout = time.Second * 60
	}
	
	if fco.Log == nil {
		return fmt.Errorf("%v Logger is nil.", op)
	}
	
	if fco.ProxyService == nil {
		return fmt.Errorf("%v Proxy Service is nil.", op)
	}
	
	return nil
	
}
