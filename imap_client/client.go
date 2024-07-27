package imap_client

import (
	"crypto/tls"
	"fmt"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"github.com/GoHippo/network/proxy/proxy_service"
	"github.com/emersion/go-imap/v2/imapclient"
	"log/slog"
	"net"
	
	"strings"
	"time"
)

type ImapClient struct {
	log *slog.Logger
	*imapclient.Client
	login bool
	proxy_jar.ProxyConfig
	Proxy               bool
	count_reconnections int
	proxyService        *proxy_service.ProxyService
}

// NewImapClient Получает imapclient c прокси или без.
func NewImapClient(addr string, dialTimeout time.Duration, block_without_proxy bool, count_reconnections int, ps *proxy_service.ProxyService) (*ImapClient, error) {
	ic := &ImapClient{count_reconnections: count_reconnections}
	
	cli, proxyConfig, err := ps.GetProxyImap(addr, dialTimeout)
	if err != nil {
		
		if err == proxy_service.ERR_JAR_PROXY_NULL && block_without_proxy {
			return nil, fmt.Errorf("Ошибка. Прокси закончились. Стоит запрет на использование сети без прокси.")
		}
		
		if err == proxy_service.ERR_JAR_PROXY_NULL {
			
			return getImapclientNotProxy(addr, dialTimeout, count_reconnections, ps)
		}
		
		if CheckErrNetwork(ic.log, err) {
			ps.DeleteProxy(proxyConfig)
			return NewImapClient(addr, dialTimeout, block_without_proxy, count_reconnections, ps)
		}
		
		return nil, err
	}
	
	ic.Client = cli
	ic.ProxyConfig = proxyConfig
	ic.Proxy = true
	
	return ic, err
}

func getImapclientNotProxy(addrMail string, dialTimeout time.Duration, count_reconnections int, ps *proxy_service.ProxyService) (*ImapClient, error) {
	// Создаем Dialer с настроенным тайм-аутом
	dialer := &net.Dialer{
		Timeout: dialTimeout,
		// KeepAlive: 0,
	}
	
	conn, err := tls.DialWithDialer(dialer, "tcp", addrMail, &tls.Config{
		NextProtos:         []string{"imap"},
		InsecureSkipVerify: true,
	})
	
	if err != nil {
		return nil, err
	}
	
	cli := imapclient.New(conn, nil)
	
	return &ImapClient{Client: cli, count_reconnections: count_reconnections, Proxy: false, proxyService: ps}, nil
}

var (
	ArrErrInternet = []string{
		`connectex: An attempt was made to access a socket in a way forbidden by its access permissions.`,
		`connectex: A socket operation was attempted to an unreachable host.`,
		`connectex: No connection`,
		`connectex: `,
		`i/o timeout`,
		`An existing connection was forcibly closed by the remote host`,
		// `no such host`,
	}
)

func (ic *ImapClient) SetIsLogin(b bool) {
	ic.login = b
}

func CheckErrNetwork(log *slog.Logger, err error) bool {
	for _, e := range ArrErrInternet {
		if strings.Contains(err.Error(), e) {
			log.Error("[CheckErrInternet] " + err.Error())
			return true
		}
	}
	return false
	// l.Error("")
}

// доработать закрытие клиента
func (ic *ImapClient) Close() {
	defer recoverLogout(ic.log)
	
	if ic.login {
		ic.Client.Logout()
	}
	ic.Client.Close()
	
	if ic.Proxy {
		ic.proxyService.FreeProxy(ic.ProxyConfig)
		ic.Proxy = false
	}
}

func recoverLogout(log *slog.Logger, ) {
	err := recover()
	if err != nil {
		log.Error(fmt.Sprintf("При закрытии клиента ошибка LOGOUT Err:%v", err))
	}
}
