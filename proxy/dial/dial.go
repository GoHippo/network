package dial

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"h12.io/socks"
	"net"
	"network/proxy/proxy_jar"
	"strings"
	"time"
)

// Создает dial - прокси для клиента fasthttp
func CreateDialFasthttp(pc proxy_jar.ProxyConfig, dialTimeout time.Duration) fasthttp.DialFunc {
	fSocksDial := func() fasthttp.DialFunc {
		dialSocks := socks.Dial(fmt.Sprintf("%v?timeout=%vs", pc.Addr, int(dialTimeout.Seconds())))
		return func(addr string) (net.Conn, error) {
			return dialSocks("tcp", addr)
		}
	}
	
	if pc.Scheme == proxy_jar.O_HTTPS {
		addr := strings.TrimLeft(pc.Addr, string(pc.Scheme)+"://")
		return fasthttpproxy.FasthttpHTTPDialerTimeout(addr, dialTimeout)
	}
	
	return fSocksDial()
}

// Создает dial - прокси для клиента imapclient.
// TODO: подумать и может сделать, чтобы возвращал чисто conn.
func CreateImapDial(pc proxy_jar.ProxyConfig, dialTimeout time.Duration, addrMail string) (*imapclient.Client, error) {
	
	dial := socks.Dial(fmt.Sprintf("%v?timeout=%vs", pc.Addr, int(dialTimeout.Seconds())))
	
	conn, err := dial("tcp", addrMail) // dial(p)
	if err != nil {
		return nil, err
	}
	
	conn = tls.Client(conn, &tls.Config{
		NextProtos:         []string{"imap"},
		InsecureSkipVerify: true,
	})
	
	return imapclient.New(conn, nil), nil
}
