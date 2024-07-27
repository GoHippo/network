package checker_proxy

import (
	"log/slog"
	"network/proxy/proxy_jar"
	"network/proxy/proxy_service"
	"sync"
	"time"
)

type loaderProxyTest struct {
	proxy proxy_jar.ProxyConfig
}

type proxyTestService struct {
	log          *slog.Logger
	proxyService *proxy_service.ProxyService
	threads      int
	wg           *sync.WaitGroup
	isBar        bool
	barAdd       func()
	loader       chan loaderProxyTest
	signalExit   chan struct{}
	good         int
	bad          int
}

type BarProxyCheck struct {
	Start func()
	End   func()
	Add   func(i int)
}

// Проверяет указынные ProxyConfig и добавляет в ProxyService и потом в файл, там проверяет на повторы
func newProxyListCheck(threads int, dialTimeout time.Duration, listProxy []proxy_jar.ProxyConfig, isBar bool, bar *BarProxyCheck, ps *proxy_service.ProxyService) (good, bad int) {
	var barAdd func()
	if isBar {
		barAdd = func() {
			bar.Add(len(listProxy))
		}
		bar.Start()
		defer bar.End()
	}
	
	pts := &proxyTestService{
		proxyService: ps,
		threads:      threads,
		isBar:        isBar,
		barAdd:       barAdd,
		wg:           &sync.WaitGroup{},
		loader:       make(chan loaderProxyTest, len(listProxy)),
		signalExit:   make(chan struct{}),
	}
	
	go pts.goPool(dialTimeout)
	defer pts.close()
	
	for _, p := range listProxy {
		pts.wg.Add(1)
		pts.loader <- loaderProxyTest{p}
	}
	
	pts.wg.Wait()
	return pts.good, pts.bad
}

func (pts *proxyTestService) goPool(dialTimeout time.Duration) {
	
	for _ = range pts.threads {
		go func() {
			for {
				select {
				case load := <-pts.loader:
					p := load.proxy
					
					if work, isImap := CheckProxyConfig(p, dialTimeout); work {
						p.IsImapSupport = isImap
						pts.proxyService.AddProxy(p)
						pts.good++
					} else {
						pts.proxyService.DeleteProxy(p)
						pts.bad++
					}
					
					pts.wg.Done()
					
					if pts.isBar {
						pts.barAdd()
					}
				
				case <-pts.signalExit:
					return
				
				default:
					time.Sleep(time.Millisecond)
				}
			}
		}()
		time.Sleep(time.Millisecond * 5)
	}
}

func (pts *proxyTestService) close() {
	for _ = range pts.threads {
		pts.signalExit <- struct{}{}
	}
	close(pts.signalExit)
	close(pts.loader)
}
