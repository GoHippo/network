package proxy_service

import (
	"fmt"
	"github.com/GoHippo/network/proxy/proxy_service/config"
	"github.com/davecgh/go-spew/spew"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestProxyService(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ps := NewProxyService(log, 2)

	testProxyServiceAdd(t, ps)

	testProxyServiceDelete(t, ps)

	testProxyServiceNull(t, ps)

	testProxyServiceDialFasthttp(t, ps)

	testRateLimit(t, ps)
	testRateLimit2(t, ps)

}

func testProxyServiceAdd(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceAdd`

	arr := []string{"https://127.0.0.1:5555", "https://127.0.0.1:6666", "socks4://127.0.0.1:7777"}
	count, err := ps.AddProxyFromArr(arr)
	if err != nil {
		t.Error(fmt.Errorf("%v: AddProxyFromArr", op))
		return
	}

	if count != len(arr) || ps.GetCountProxy() != 3 {
		t.Error(fmt.Errorf("%v: AddProxyFromArr: count!=3 ", op))
		return
	}

}

func testProxyServiceDelete(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceDelete`

	for _ = range 3 {
		proxy, err := ps.GetProxy()
		if err != nil {
			t.Error(fmt.Errorf("%v: GetProxy: %w", op, err))
			return
		}
		ps.DeleteProxy(proxy)
	}

	if ps.GetCountProxy() != 0 {
		t.Error(fmt.Errorf("%v", op))
		return
	}

}

func testProxyServiceNull(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceNull`
	if ps.GetCountProxy() != 0 {
		testProxyServiceDelete(t, ps)
	}

	_, config, err := ps.GetFasthttpProxy(time.Second * 15)
	if err != ERR_JAR_PROXY_NULL {
		t.Errorf("%v: ERR_JAR_PROXY_NULL not work:%v", op, err)
	}
	ps.FreeProxy(config)
}

func testProxyServiceDialFasthttp(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceDialFasthttp`

	if ps.GetCountProxy() != 0 {
		testProxyServiceDelete(t, ps)
	}
	testProxyServiceAdd(t, ps)

	fFree := func(config config.ProxyConfig) {
		time.Sleep(2 * time.Second)
		ps.FreeProxy(config)
	}

	dial, config, err := ps.GetFasthttpProxy(time.Second * 15)
	if err != nil || dial == nil || config.Addr == "" {
		t.Errorf("%v:  with fasthttpdial not work:%v", op, err)
	}
	fFree(config)

	dial, config, err = ps.GetFasthttpProxy(time.Second * 15)
	if err != nil || dial == nil || config.Addr == "" {
		t.Errorf("%v:  with fasthttpdial not work:%v", op, err)
	}
	fFree(config)

	//_, config, err = ps.GetProxyImap("imap.outlook.com:993", time.Second*15)
	//if !strings.Contains(err.Error(), "connectex") {
	//	t.Errorf("%v:  with imapdial not work:%v", op, err)
	//}

	ps.DeleteProxy(config)

	//_, _, err = ps.GetProxyImap("imap.outlook.com:993", time.Second*15)
	//if err != ERR_JAR_PROXY_NULL {
	//	t.Errorf("%v:  with imapdial not work:%v", op, err)
	//}

}

func testRateLimit(t *testing.T, ps *ProxyService) {
	var op = `testRateLimit`
	testProxyServiceAdd(t, ps)

	var arrConfig []config.ProxyConfig

	go func() {
		time.Sleep(1 * time.Second)
		ps.FreeProxy(arrConfig[0])
	}()

	for _ = range 7 {
		p, err := ps.GetProxy()
		if err != nil {
			t.Error(fmt.Errorf("%v: %w", op, err))
		}
		arrConfig = append(arrConfig, p)
	}

	for _, r := range ps.jar {
		if r != 2 {
			t.Error(fmt.Errorf("%v", op))
			spew.Dump(ps.jar)
		}
	}

	for i := range 3 {
		ps.DeleteProxy(arrConfig[i])
	}

	if ps.GetCountProxy() != 0 {
		t.Error(fmt.Errorf("%v", op))
	}
}

func testRateLimit2(t *testing.T, ps *ProxyService) {
	var op = `testRateLimit`
	testProxyServiceAdd(t, ps)

	var arrConfig []config.ProxyConfig

	go func() {
		time.Sleep(1 * time.Second)
		ps.FreeProxy(arrConfig[0])
	}()

	wg := &sync.WaitGroup{}
	for _ = range 7 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := ps.GetProxy()
			if err != nil {
				t.Error(fmt.Errorf("%v: %w", op, err))
			}
			arrConfig = append(arrConfig, p)
		}()

	}

	wg.Wait()

	for _, r := range ps.jar {
		if r != 2 {
			t.Error(fmt.Errorf("%v", op))
			spew.Dump(ps.jar)
		}
	}

	for i := range 3 {
		ps.DeleteProxy(arrConfig[i])
	}

	if ps.GetCountProxy() != 0 {
		t.Error(fmt.Errorf("%v", op))
	}
}
