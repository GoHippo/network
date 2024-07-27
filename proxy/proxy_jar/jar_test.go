package proxy_jar

import (
	"fmt"
	"strconv"
	"testing"
)

func TestProxyJar(t *testing.T) {
	
	jar := &JarProxy{}
	testProxyJarAdd(jar, t)
	testProxyJarDelete(jar, t)
}

func testProxyJarAdd(jar *JarProxy, t *testing.T) {
	var op = `TestProxyJarAdd`
	
	for i := range 5 {
		
		var scheme ProxyScheme
		var isImap bool
		switch i%2 == 0 {
		case true:
			scheme = O_HTTPS
		
		case false:
			scheme = O_SOCKS4
			isImap = true
		}
		host := "127.0.0.1:808" + strconv.Itoa(i)
		
		jar.AddProxy(ProxyConfig{
			Addr:          fmt.Sprintf("%s://%s", scheme, host),
			Scheme:        scheme,
			Host:          host,
			IsImapSupport: isImap,
		})
		
	}
	
	if jar.GetCountProxy() != 5 {
		t.Errorf("%v: Proxy not add", op)
	}
	
	if jar.GetCountImapProxy() != 2 {
		t.Errorf("%v: ProxyImap not add", op)
	}
}

func testProxyJarDelete(jar *JarProxy, t *testing.T) {
	var op = `TestProxyJarDelete`
	
	jar.DeleteProxy(ProxyConfig{
		Addr:          "https://127.0.0.1:8082",
		Scheme:        O_HTTPS,
		Host:          "127.0.0.1:8082",
		IsImapSupport: false,
	})
	
	if jar.GetCountProxy() != 4 {
		t.Errorf("%v: Proxy not delete", op)
	}
}
