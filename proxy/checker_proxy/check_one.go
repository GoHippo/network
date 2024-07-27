package checker_proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/GoHippo/network/proxy/dial"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"github.com/valyala/fasthttp"
	
	"time"
)

// CheckProxyConfig CheckProxy проверяет proxy_jar.ProxyConfig на валид fasthttp и imap.
func CheckProxyConfig(config proxy_jar.ProxyConfig, dialTimeout time.Duration) (bool, bool) {
	client := &fasthttp.Client{
		Dial:      dial.CreateDialFasthttp(config, dialTimeout),
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer client.CloseIdleConnections()
	
	isWork := CheckTreeDomains(client, dialTimeout) == nil
	isImap := CheckImapConfig(config, dialTimeout) == nil
	
	return isWork, isImap
}

// CheckTreeDomains Проверяет клиент с прокси по трем доменам, если во всех ошибка, то вернет ошибку.
func CheckTreeDomains(client *fasthttp.Client, dialTimeout time.Duration) error {
	// https://jsonip.com
	// https://api.seeip.org/jsonip
	// https://icanhazip.com
	
	if s, _, err := client.GetTimeout(nil, "https://b.cdnst.net/javascript/ads/ad.js", dialTimeout); err == nil || s == 200 || s == 302 {
		return nil
	}
	
	if s, _, err := client.GetTimeout(nil, "https://jsonip.com", dialTimeout); err == nil || s == 200 || s == 302 {
		return nil
	}
	
	if s, _, err := client.GetTimeout(nil, "https://api.seeip.org/jsonip", dialTimeout); err == nil || s == 200 || s == 302 {
		return nil
	} else {
		return fmt.Errorf("the proxy failed verification:%w", err)
	}
}

// CheckImapConfig Проверяет proxy_jar.ProxyConfig на валид imap.
func CheckImapConfig(config proxy_jar.ProxyConfig, dialTimeout time.Duration) error {
	if config.Scheme == proxy_jar.O_HTTPS {
		return fmt.Errorf("HTTPS не поддерживает IMAP")
	}
	
	cli, err := dial.CreateImapDial(config, dialTimeout, "outlook.office365.com:993")
	if err != nil {
		return err
	}
	
	defer cli.Close()
	return nil
}
