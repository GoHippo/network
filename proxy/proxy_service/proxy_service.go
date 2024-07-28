package proxy_service

import (
	"fmt"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"github.com/GoHippo/slogpretty/sl"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/valyala/fasthttp"
	"os"
	
	dial_service "github.com/GoHippo/network/proxy/dial"
	"log/slog"
	"net/url"
	
	"slices"
	"strings"
	"time"
)

func NewProxyService(log *slog.Logger) *ProxyService {
	
	ps := &ProxyService{log: log, jar: &proxy_jar.JarProxy{}, loader: make(chan poolloader)}
	ps.goPool()
	
	ps.log.Debug(fmt.Sprintf("Создан ProxyService: jar=%v", len(ps.jar.ArrConfig)))
	
	return ps
}

func (ps *ProxyService) goPool() {
	var arrUsedProxyConfig []proxy_jar.ProxyConfig
	
	go func() {
		for {
			
			select {
			case load := <-ps.loader:
				
				// l.Debug(fmt.Sprintf("получен лоадер ProxyService:%v %v", load.cmd, load.proxy))
				
				switch load.cmd {
				
				case ADD:
					ps.jar.AddProxy(load.proxy)
					load.resp <- poolloader{}
				
				case DELETE:
					ps.jar.DeleteProxy(load.proxy)
					load.resp <- poolloader{}
				
				case GET:
					
					fGet := func() bool {
						
						if ps.jar.GetCountProxy() == 0 {
							load.resp <- poolloader{cmd: ERR_NULL}
							return true
						}
						
						for _, p := range ps.jar.ArrConfig {
							if !slices.Contains(arrUsedProxyConfig, p) {
								// ps.jar.ArrConfig[i].IsUsed = true
								arrUsedProxyConfig = append(arrUsedProxyConfig, p)
								load.resp <- poolloader{proxy: p}
								return true
							}
						}
						return false
					}
					
					if !fGet() {
						load.resp <- poolloader{cmd: WAIT_FREE}
					}
				
				case GET_IMAP:
					// возможен пиздец, из-за того что проходит из другого потока в обратку, где должен быть результат
					fGet := func() bool {
						if ps.jar.GetCountImapProxy() == 0 {
							load.resp <- poolloader{cmd: ERR_NULL}
							return true
						}
						
						for _, p := range ps.jar.ArrConfig {
							if !slices.Contains(arrUsedProxyConfig, p) && p.IsImapSupport {
								arrUsedProxyConfig = append(arrUsedProxyConfig, p)
								load.resp <- poolloader{proxy: p}
								return true
							}
						}
						return false
					}
					
					if !fGet() {
						load.resp <- poolloader{cmd: WAIT_FREE}
					}
				
				case FREE:
					for _, p := range ps.jar.ArrConfig {
						if slices.Contains(arrUsedProxyConfig, p) {
							arrUsedProxyConfig = slices.DeleteFunc(arrUsedProxyConfig, func(c proxy_jar.ProxyConfig) bool {
								return c == p
							})
							break
						}
					}
					load.resp <- poolloader{}
				
				case CLOSE:
					close(ps.loader)
					return
					
				}
			default:
				time.Sleep(time.Millisecond)
			}
			
		}
		
	}()
}

func (ps *ProxyService) AddProxy(p proxy_jar.ProxyConfig) {
	loader := poolloader{cmd: ADD, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)
	
	ps.loader <- loader
	<-loader.resp
	
}

func (ps *ProxyService) DeleteProxy(p proxy_jar.ProxyConfig) {
	
	loader := poolloader{cmd: DELETE, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)
	
	ps.loader <- loader
	<-loader.resp
}

func (ps *ProxyService) FreeProxy(p proxy_jar.ProxyConfig) {
	loader := poolloader{cmd: FREE, proxy: p, resp: make(chan poolloader)}
	defer close(loader.resp)
	
	ps.loader <- loader
	<-loader.resp
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

// ====================== AddToJar ======================

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

func (ps *ProxyService) AddProxtFromFile(path string, scheme string) (count int, err error) {
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

func (ps *ProxyService) ConvertStrToProxyConfig(arr []string) (arrConfig []proxy_jar.ProxyConfig, err error) {
	var op = `network.proxy.proxy_service.ConvertStrToProxyConfig`
	
	for _, str := range arr {
		str = strings.TrimSpace(str)
		u, err := url.Parse(str)
		if err != nil {
			err = fmt.Errorf("Error parsing proxy [%s]:%w ", str, err)
			ps.log.Error(op, sl.Err(err))
			return arrConfig, err
		}
		
		sc := proxy_jar.ProxyScheme(u.Scheme)
		if sc == proxy_jar.O_HTTPS || sc == proxy_jar.O_SOCKS4 || sc == proxy_jar.O_SOCKS5 || sc == proxy_jar.O_SOCKS4a {
			proxyConfig := proxy_jar.ProxyConfig{Scheme: sc, Addr: str, Host: u.Host}
			
			if sc == proxy_jar.O_SOCKS4 || sc == proxy_jar.O_SOCKS5 || sc == proxy_jar.O_SOCKS4a {
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
func (ps *ProxyService) GetFasthttpProxy(dialTimeout time.Duration) (fasthttp.DialFunc, proxy_jar.ProxyConfig, error) {
	loader := poolloader{cmd: GET, resp: make(chan poolloader)}
	defer close(loader.resp)
	
	for {
		ps.loader <- loader
		load := <-loader.resp
		
		if load.cmd == ERR_NULL {
			return nil, proxy_jar.ProxyConfig{}, ERR_JAR_PROXY_NULL
		}
		
		if load.cmd == WAIT_FREE {
			time.Sleep(time.Second)
			continue
		}
		
		dial := dial_service.CreateDialFasthttp(load.proxy, dialTimeout)
		
		return dial, load.proxy, nil
	}
}

// GetProxyImap - сразу будет ошибка, если прокси не валидный.
// Если ошибка в прокси, то он не освобождает прокси и он не валид.
func (ps *ProxyService) GetProxyImap(addrMail string, dialTimeout time.Duration) (*imapclient.Client, proxy_jar.ProxyConfig, error) {
	for {
		loader := poolloader{cmd: GET_IMAP, resp: make(chan poolloader)}
		
		ps.loader <- loader
		
		load := <-loader.resp
		defer close(loader.resp)
		
		if load.cmd == ERR_NULL {
			return nil, proxy_jar.ProxyConfig{}, ERR_JAR_PROXY_NULL
		}
		
		if load.cmd == WAIT_FREE {
			time.Sleep(time.Second)
			continue
		}
		
		imapclient, err := dial_service.CreateImapDial(load.proxy, dialTimeout, addrMail)
		
		return imapclient, load.proxy, err
	}
}

// ====================== Count  ======================

func (ps *ProxyService) GetCountProxy() int {
	return ps.jar.GetCountProxy()
}

func (ps *ProxyService) GetCountProxyImap() int {
	return ps.jar.GetCountImapProxy()
}
