package checker_proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/GoHippo/network/proxy/config"
	"github.com/valyala/fasthttp"
	"h12.io/socks"
	"net"
	"net/url"
	"strings"
	"time"
)

// TODO: сделать аутификацию юзер пасс сюда. Добавить возможность загружать уни.
// TODO: добавить возможность проверять прокси при получении из сервиса и сразу заменить на другой.
// TODO: разобратся как можно добавлять прокси заголовки и что делают
func CheckProxyUni(ipport string, dialTimeout time.Duration) (config.ProxyConfig, bool) {

	ipport = strings.TrimSpace(ipport)

	if socks4(ipport, dialTimeout) {
		addr := fmt.Sprintf("%v://%v", config.O_SOCKS4, ipport)
		u, _ := url.Parse(addr)
		return config.ProxyConfig{
			Addr:          addr,
			Scheme:        config.O_SOCKS4,
			Host:          u.Host,
			IsImapSupport: true,
		}, true
	}

	if socks4a(ipport, dialTimeout) {
		addr := fmt.Sprintf("%v://%v", config.O_SOCKS4a, ipport)
		u, _ := url.Parse(addr)
		return config.ProxyConfig{
			Addr:          addr,
			Scheme:        config.O_SOCKS4a,
			Host:          u.Host,
			IsImapSupport: true,
		}, true
	}

	if socks5(ipport, dialTimeout) {
		addr := fmt.Sprintf("%v://%v", config.O_SOCKS5, ipport)
		u, _ := url.Parse(addr)
		return config.ProxyConfig{
			Addr:          addr,
			Scheme:        config.O_SOCKS5,
			Host:          u.Host,
			IsImapSupport: true,
		}, true
	}

	if https(ipport, dialTimeout) {
		addr := fmt.Sprintf("%v://%v", config.O_HTTPS, ipport)
		u, _ := url.Parse(addr)
		return config.ProxyConfig{
			Addr:          addr,
			Scheme:        config.O_HTTPS,
			Host:          u.Host,
			IsImapSupport: true,
		}, true
	}

	if httpJust(ipport, dialTimeout) {
		addr := fmt.Sprintf("%v://%v", config.O_HTTP, ipport)
		u, _ := url.Parse(addr)
		return config.ProxyConfig{
			Addr:          addr,
			Scheme:        config.O_HTTP,
			Host:          u.Host,
			IsImapSupport: true,
		}, true
	}

	return config.ProxyConfig{}, false

}

func CheckProxyIpPort(ipport string, schema config.ProxyScheme, dialTimeout time.Duration) (config.ProxyConfig, bool) {

	ipport = strings.TrimSpace(ipport)

	switch schema {
	case config.O_SOCKS4:
		if socks4(ipport, dialTimeout) {
			addr := fmt.Sprintf("%v://%v", config.O_SOCKS4, ipport)
			u, _ := url.Parse(addr)
			return config.ProxyConfig{
				Addr:          addr,
				Scheme:        config.O_SOCKS4,
				Host:          u.Host,
				IsImapSupport: true,
			}, true
		}
	case config.O_SOCKS4a:
		if socks4a(ipport, dialTimeout) {
			addr := fmt.Sprintf("%v://%v", config.O_SOCKS4a, ipport)
			u, _ := url.Parse(addr)
			return config.ProxyConfig{
				Addr:          addr,
				Scheme:        config.O_SOCKS4a,
				Host:          u.Host,
				IsImapSupport: true,
			}, true
		}
	case config.O_SOCKS5:
		if socks5(ipport, dialTimeout) {
			addr := fmt.Sprintf("%v://%v", config.O_SOCKS5, ipport)
			u, _ := url.Parse(addr)
			return config.ProxyConfig{
				Addr:          addr,
				Scheme:        config.O_SOCKS5,
				Host:          u.Host,
				IsImapSupport: true,
			}, true
		}
	case config.O_HTTPS:
		if https(ipport, dialTimeout) {
			addr := fmt.Sprintf("%v://%v", config.O_HTTPS, ipport)
			u, _ := url.Parse(addr)
			return config.ProxyConfig{
				Addr:          addr,
				Scheme:        config.O_HTTPS,
				Host:          u.Host,
				IsImapSupport: true,
			}, true
		}
	case config.O_HTTP:
		if httpJust(ipport, dialTimeout) {
			addr := fmt.Sprintf("%v://%v", config.O_HTTP, ipport)
			u, _ := url.Parse(addr)
			return config.ProxyConfig{
				Addr:          addr,
				Scheme:        config.O_HTTP,
				Host:          u.Host,
				IsImapSupport: true,
			}, true
		}
	}

	return config.ProxyConfig{}, false
}

func socks4(ipport string, dialTimeout time.Duration) bool {
	dialSocks := socks.Dial(fmt.Sprintf("socks4://%v?timeout=%vs", ipport, int(dialTimeout.Seconds())))
	c, err := dialSocks("tcp", ipport)
	if err != nil {
		fmt.Printf("socks4 in %v err:%v\n", ipport, err)
		return false
	}
	defer c.Close()

	client := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return c, err
		},
	}

	return CheckTreeDomains(client, dialTimeout) == nil
}

func socks4a(ipport string, dialTimeout time.Duration) bool {
	dialSocks := socks.Dial(fmt.Sprintf("socks4a://%v?timeout=%vs", ipport, int(dialTimeout.Seconds())))
	c, err := dialSocks("tcp", ipport)
	if err != nil {
		fmt.Printf("socks4a in %v err:%v\n", ipport, err)
		return false
	}
	defer c.Close()

	client := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return c, err
		},
	}

	return CheckTreeDomains(client, dialTimeout) == nil
}

func socks5(ipport string, dialTimeout time.Duration) bool {
	dialSocks := socks.Dial(fmt.Sprintf("socks5://%v?timeout=%vs", ipport, int(dialTimeout.Seconds())))
	c, err := dialSocks("tcp", ipport)
	if err != nil {
		fmt.Printf("socks5 in %v err:%v\n", ipport, err)
		return false
	}
	defer c.Close()

	client := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return c, err
		},
	}

	return CheckTreeDomains(client, dialTimeout) == nil
}

func https(ipport string, dialTimeout time.Duration) bool {
	config := &tls.Config{
		InsecureSkipVerify: true, // Не проверяем сертификаты для примера
	}

	d := new(net.Dialer)
	d.Timeout = dialTimeout

	c, err := tls.DialWithDialer(d, "tcp", ipport, config)
	if err != nil {
		fmt.Printf("https in %v err:%v\n", ipport, err)
		return false
	}
	defer c.Close()

	client := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return c, err
		},
	}

	return CheckTreeDomains(client, dialTimeout) == nil
}

func httpJust(ipport string, dialTimeout time.Duration) bool {

	d := new(net.Dialer)
	d.Timeout = dialTimeout
	c, err := d.Dial("tcp", ipport)
	if err != nil {
		return false
	}
	defer c.Close()

	client := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return c, err
		},
	}

	//client.Dial=fasthttp.Prox

	return CheckTreeDomains(client, dialTimeout) == nil

}
