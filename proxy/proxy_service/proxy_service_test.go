package proxy_service

import (
	"MeteorMail/internal/network/proxy/proxy_jar"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProxyService(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ps := NewProxyService(log)
	
	testProxyServiceAdd(t, ps)
	testProxyServiceDelete(t, ps)
	testProxyServiceNull(t, ps)
	testProxyServiceDialFasthttp(t, ps)
	
}

func testProxyServiceAdd(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceAdd`
	
	arr := []string{"https://127.0.0.1:5555", "https://127.0.0.1:6666", "socks4://127.0.0.1:7777"}
	count, err := ps.AddProxyFromArr(arr)
	if err != nil {
		t.Error(fmt.Errorf("%v: AddProxyFromArr", op))
		return
	}
	
	if count != len(arr) || ps.jar.GetCountProxy() != 3 {
		t.Error(fmt.Errorf("%v: AddProxyFromArr: count!=3 ", op))
		return
	}
}

func testProxyServiceDelete(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceDelete`
	
	for _ = range 3 {
		ps.DeleteProxy(ps.jar.ArrConfig[0])
	}
	
	if ps.jar.GetCountProxy() != 0 {
		t.Error(fmt.Errorf("%v", op))
		return
	}
}

func testProxyServiceNull(t *testing.T, ps *ProxyService) {
	var op = `testProxyServiceNull`
	if ps.jar.GetCountProxy() != 0 {
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
	
	if ps.jar.GetCountProxy() != 0 {
		testProxyServiceDelete(t, ps)
	}
	testProxyServiceAdd(t, ps)
	
	fFree := func(config proxy_jar.ProxyConfig) {
		time.Sleep(2 * time.Second)
		ps.FreeProxy(config)
	}
	
	dial, config, err := ps.GetFasthttpProxy(time.Second * 15)
	if err != nil || dial == nil || config.Addr == "" {
		t.Errorf("%v:  with fasthttpdial not work:%v", op, err)
	}
	go fFree(config)
	
	dial, config, err = ps.GetFasthttpProxy(time.Second * 15)
	if err != nil || dial == nil || config.Addr == "" {
		t.Errorf("%v:  with fasthttpdial not work:%v", op, err)
	}
	go fFree(config)
	
	_, config, err = ps.GetProxyImap("imap.outlook.com:993", time.Second*15)
	if !strings.Contains(err.Error(), "connectex") {
		t.Errorf("%v:  with imapdial not work:%v", op, err)
	}
	
	ps.DeleteProxy(config)
	
	_, _, err = ps.GetProxyImap("imap.outlook.com:993", time.Second*15)
	if err != ERR_JAR_PROXY_NULL {
		t.Errorf("%v:  with imapdial not work:%v", op, err)
	}
	
}
