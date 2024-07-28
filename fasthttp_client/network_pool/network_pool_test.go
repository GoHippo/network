package network_pool

import (
	"SteamRecovery/pkg/console_bar"
	"encoding/json"
	"fmt"
	"github.com/GoHippo/network/fasthttp_client"
	"github.com/GoHippo/network/fasthttp_client/cookies_jar"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"github.com/GoHippo/network/proxy/proxy_service"
	"github.com/GoHippo/slogpretty/sl"
	"github.com/GoHippo/slogpretty/slogpretty"
	"github.com/valyala/fasthttp"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type Resourse struct {
	DoOpt fasthttp_client.DoOption
}

type ResTest struct {
	arrResult []string
	lock      sync.Mutex
	arrRes    []Resourse
	Log       *slog.Logger
}

func (rt *ResTest) Check(client *fasthttp_client.FasthttpClient, resource any) {
	// https://api.seeip.org/jsonip {"ip":"85.192.63.92"}
	var req = fasthttp.AcquireRequest()
	var resp = fasthttp.AcquireResponse()
	
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	
	res := resource.(Resourse)
	
	req.SetRequestURI("https://api.seeip.org/jsonip")
	
	_, err := client.Do(req, resp, res.DoOpt)
	if err != nil {
		rt.Log.Error("Error do request", sl.Err(err))
		return
	}
	
	data := struct {
		Ip string `json:"ip"`
	}{}
	
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		rt.Log.Error("Error unmarshal json", sl.Err(err))
		return
	}
	
	if data.Ip != "" {
		rt.WriteResult(data.Ip)
	}
}

func (rt *ResTest) WriteResult(s string) {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	rt.arrResult = append(rt.arrResult, s)
}

func (rt *ResTest) LenResource() int {
	return len(rt.arrRes)
}

func (rt *ResTest) GetResource() any {
	res := rt.arrRes[0]
	rt.arrRes = rt.arrRes[1:]
	return res
}

func TestNewNetworkPool(t *testing.T) {
	log := slogpretty.SetupPrettySlog(slog.LevelInfo)
	
	rt := &ResTest{
		lock: sync.Mutex{},
		Log:  log,
	}
	
	for i := range 1 {
		rt.arrRes = append(rt.arrRes, Resourse{DoOpt: fasthttp_client.DoOption{
			ID:                  fmt.Sprintf("%v", i),
			DoCountReconnecting: 2,
			NotBodyDecode:       false,
			Jar:                 cookies_jar.NewJar(log),
			ErrCounter:          nil,
		}})
	}
	
	bar := console_bar.NewConsoleBar("TestNewNetworkPool")
	
	ps := proxy_service.NewProxyService(log)
	ps.AddProxy(proxy_jar.ProxyConfig{
		Addr:          "https://127.0.0.1:6666",
		Scheme:        "https",
		Host:          "127.0.0.1:6666",
		IsImapSupport: false,
	})
	
	NewNetworkPool(NetworkPoolOptions{
		ActionBox: ActionBox(rt),
		CliOptions: fasthttp_client.FastHttpClientOptions{
			DialTimeout:              time.Second * 5,
			MaxConnsPerHost:          5,
			Log:                      log,
			ProxyService:             ps,
			ProxyUse:                 true,
			BlockNetworkWithoutProxy: false,
			CountReconnections:       0,
		},
		Threads:        5,
		Log:            log,
		FuncSignalDone: bar.Add,
	})
	bar.Close("end")
	
	fmt.Println(rt.arrResult)
	
}
