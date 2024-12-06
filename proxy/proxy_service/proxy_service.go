package proxy_service

import (
	"fmt"
	"github.com/GoHippo/network/proxy/checker_proxy"
	"github.com/GoHippo/network/proxy/config"
	"github.com/GoHippo/slogpretty/sl"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/valyala/fasthttp"
	"maps"
	"os"
	"strconv"

	dial_service "github.com/GoHippo/network/proxy/dial"
	"log/slog"
	"net/url"

	"strings"
	"time"
)

func NewProxyService(log *slog.Logger, conn_limit int, timeout_dial time.Duration) *ProxyService {
	if conn_limit < 1 {
		conn_limit = 1
	}

	if timeout_dial == 0 {
		timeout_dial = 15 * time.Second
	}

	ps := &ProxyService{
		log:         log,
		loader:      make(chan poolloader),
		timeoutDial: timeout_dial,
		conn_limit:  conn_limit,
		jar:         make(map[config.ProxyConfig]int),
		jarRetries:  make(map[config.ProxyConfig]int),
	}
	ps.goPool()

	return ps
}

func (ps *ProxyService) goPool() {

	go func() {
		for {
			load := <-ps.loader

			switch load.cmd {

			case ADD:

				if _, ok := ps.jar[load.proxy]; !ok {
					ps.jar[load.proxy] = 0
				}
				load.resp <- poolloader{}

			case DELETE:
				maps.DeleteFunc(ps.jar, func(config config.ProxyConfig, i int) bool {
					return config.Addr == load.proxy.Addr
				})
				load.resp <- poolloader{}

			case GET:

				if len(ps.jar) == 0 && len(ps.jarRetries) != 0 {
					ps.jar = ps.jarRetries
					ps.jarRetries = make(map[config.ProxyConfig]int)
				}

				if len(ps.jar) == 0 {
					load.resp <- poolloader{cmd: ERR_NULL}
					continue
				}

				fGet := func() bool {

					var config config.ProxyConfig
					var rate_limit int

					for p, i := range ps.jar {
						if ps.conn_limit > i {
							if i == 0 || i < rate_limit || config.Addr == "" {
								config = p
								rate_limit = i
							}
						}
					}

					if config.Addr != "" {
						ps.jar[config] = rate_limit + 1
						load.resp <- poolloader{proxy: config}
						return true
					}
					return false
				}

				if !fGet() {
					load.resp <- poolloader{cmd: WAIT_FREE}
				}

			case FREE:

				if r, ok := ps.jar[load.proxy]; ok {
					if r != 0 {
						r--
					}
					ps.jar[load.proxy] = r
				}
				load.resp <- poolloader{}

			case FREE_RETRIES:
				if _, ok := ps.jarRetries[load.proxy]; !ok {
					ps.jarRetries[load.proxy] = 0
				}
				load.resp <- poolloader{}

			case COUNT:
				load.resp <- poolloader{cmd: cmd_poolloader(strconv.Itoa(len(ps.jar)))}

			case CLOSE:
				close(ps.loader)
				return

			}
		}

	}()
}

func (ps *ProxyService) AddProxy(p config.ProxyConfig) {
	loader := poolloader{cmd: ADD, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)

	ps.loader <- loader
	<-loader.resp

}

func (ps *ProxyService) GetProxy() (config.ProxyConfig, error) {
	loader := poolloader{cmd: GET, resp: make(chan poolloader)}
	defer close(loader.resp)

	for {
		ps.loader <- loader
		load := <-loader.resp

		if load.cmd == ERR_NULL {
			return config.ProxyConfig{}, ERR_JAR_PROXY_NULL
		}

		if load.cmd == WAIT_FREE {
			time.Sleep(time.Second)
			continue
		}

		return load.proxy, nil
	}

}

func (ps *ProxyService) GetProxyWithValidate() (config.ProxyConfig, error) {
	loader := poolloader{cmd: GET, resp: make(chan poolloader)}
	defer close(loader.resp)

	for {
		ps.loader <- loader
		load := <-loader.resp

		if load.cmd == ERR_NULL {
			return config.ProxyConfig{}, ERR_JAR_PROXY_NULL
		}

		if load.cmd == WAIT_FREE {
			time.Sleep(time.Second)
			continue
		}

		if !checker_proxy.CheckProxyConfig(load.proxy, ps.timeoutDial) {
			ps.DeleteProxy(load.proxy)
			continue
		}

		return load.proxy, nil
	}

}

func (ps *ProxyService) DeleteProxy(p config.ProxyConfig) {

	loader := poolloader{cmd: DELETE, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)

	ps.loader <- loader
	<-loader.resp
}

func (ps *ProxyService) FreeProxy(p config.ProxyConfig) {
	loader := poolloader{cmd: FREE, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)

	ps.loader <- loader
	<-loader.resp
}

func (ps *ProxyService) FreeRetriesProxy(p config.ProxyConfig) {
	ps.DeleteProxy(p)

	loader := poolloader{cmd: FREE_RETRIES, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)

	ps.loader <- loader
	<-loader.resp
}

func (ps *ProxyService) GetCountProxy() int {
	loader := poolloader{cmd: COUNT, resp: make(chan poolloader)}
	defer close(loader.resp)

	ps.loader <- loader
	resp := <-loader.resp
	i, _ := strconv.Atoi(string(resp.cmd))
	return i
}

func (ps *ProxyService) Close() {
	ps.loader <- poolloader{cmd: CLOSE}
}

/*func (ps *ProxyService) AddProxyFromFile(filePath string, isBar bool, bar *BarProxyCheck) (good, bad int) {
	arrProxyConfig, err := getProxyConfigWithFilePath(filePath)
	if err != nil {
		ps.log.Error("Ошибка. При добавлении конфигов прокси с файла: " + err.Error())
		return 0, 0
	}

	if len(arrProxyConfig) == 0 {
		return 0, 0
	}

	return newProxyListCheck(arrProxyConfig, isBar, bar)
}*/

// ====================== Add ======================

func (ps *ProxyService) AddProxyFromArr(arr []string) (count int, err error) {
	arrProxyConfig, err := ps.ConvertStrToProxyConfig(arr)
	if err != nil {
		return 0, err
	}

	if len(arrProxyConfig) == 0 {
		return 0, err
	}

	for _, p := range arrProxyConfig {
		ps.AddProxy(p)
	}

	count = len(arrProxyConfig)
	return count, nil
}

func (ps *ProxyService) AddProxyFromFile(path string, scheme string) (count int, err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		ps.log.Error("Error reading file proxy", sl.Err(err))
		return 0, err
	}

	var arr []string
	for _, line := range strings.Split(string(file), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		sp := strings.Split(line, ":")

		switch len(sp) {
		case 2:
			arr = append(arr, fmt.Sprintf("%v://%v:%v", scheme, sp[0], sp[1]))
		case 4:
			arr = append(arr, fmt.Sprintf("%v://%v:%v@%v:%v", scheme, sp[2], sp[3], sp[0], sp[1]))
		default:
			return 0, fmt.Errorf("Format proxy err. (ip:port or ip:port:user:pass)")
		}
	}
	return ps.AddProxyFromArr(arr)
}

func (ps *ProxyService) ConvertStrToProxyConfig(arr []string) (arrConfig []config.ProxyConfig, err error) {
	var op = `network.proxy.proxy_service.ConvertStrToProxyConfig`

	for _, str := range arr {
		str = strings.TrimSpace(str)
		u, err := url.Parse(str)
		if err != nil {
			err = fmt.Errorf("Error parsing proxy [%s]:%w ", str, err)
			ps.log.Error(op, sl.Err(err))
			return arrConfig, err
		}

		sc := config.ProxyScheme(u.Scheme)
		if sc == config.O_HTTP || sc == config.O_HTTPS || sc == config.O_SOCKS4 || sc == config.O_SOCKS5 || sc == config.O_SOCKS4a {
			proxyConfig := config.ProxyConfig{Scheme: sc, Addr: str, Host: u.Host}

			if sc == config.O_SOCKS4 || sc == config.O_SOCKS5 || sc == config.O_SOCKS4a {
				proxyConfig.IsImapSupport = true
			}

			arrConfig = append(arrConfig, proxyConfig)
		} else {
			err = fmt.Errorf("Error parsing scheme proxy [%s]:%w ", str, err)
			ps.log.Error(op, sl.Err(err))
			return arrConfig, err
		}

	}

	ps.log.Debug("Add proxy", slog.Int("count", len(arrConfig)))

	return arrConfig, nil
}

// ====================== ClientDialProxy ======================

// GetFasthttpProxy не проверяет прокси - это сделано для экономии ресурсов сети. Только выдает с нужным fasthttp.DialFunc
func (ps *ProxyService) GetFasthttpProxy(dialTimeout time.Duration) (fasthttp.DialFunc, config.ProxyConfig, error) {
	loader := poolloader{cmd: GET, resp: make(chan poolloader)}
	defer close(loader.resp)

	for {
		ps.loader <- loader
		load := <-loader.resp

		if load.cmd == ERR_NULL {
			return nil, config.ProxyConfig{}, ERR_JAR_PROXY_NULL
		}

		if load.cmd == WAIT_FREE {
			time.Sleep(time.Second)
			continue
		}

		if !checker_proxy.CheckProxyConfig(load.proxy, ps.timeoutDial) {
			ps.DeleteProxy(load.proxy)
			continue
		}

		dial := dial_service.CreateDialFasthttp(load.proxy, dialTimeout)

		return dial, load.proxy, nil
	}
}

// GetProxyImap - сразу будет ошибка, если прокси не валидный.
// Если ошибка в прокси, то он не освобождает прокси и он не валид.
func (ps *ProxyService) GetProxyImap(addrMail string, dialTimeout time.Duration) (*imapclient.Client, config.ProxyConfig, error) {
	for {
		loader := poolloader{cmd: GET_IMAP, resp: make(chan poolloader)}

		ps.loader <- loader

		load := <-loader.resp
		defer close(loader.resp)

		if load.cmd == ERR_NULL {
			return nil, config.ProxyConfig{}, ERR_JAR_PROXY_NULL
		}

		if load.cmd == WAIT_FREE {
			time.Sleep(time.Second)
			continue
		}

		imapclient, err := dial_service.CreateImapDial(load.proxy, dialTimeout, addrMail)

		return imapclient, load.proxy, err
	}
}
