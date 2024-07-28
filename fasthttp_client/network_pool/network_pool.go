package network_pool

import (
	"github.com/GoHippo/network/fasthttp_client"
	"log"
	"log/slog"
	"sync"
)

type ActionBox interface {
	Check(client *fasthttp_client.FasthttpClient, resource any)
	LenResource() int
	GetResource() any
}

type NetworkPoolOptions struct {
	ActionBox      ActionBox
	CliOptions     fasthttp_client.FastHttpClientOptions
	Threads        int
	Log            *slog.Logger
	FuncSignalDone func(i int)
}

type loader_resource struct {
	res any
}

type NetworkPool struct {
	NetworkPoolOptions
	loader       chan loader_resource
	wg           *sync.WaitGroup
	clientStatic *fasthttp_client.FasthttpClient
}

func NewNetworkPool(opt NetworkPoolOptions) {
	if opt.Threads == 0 {
		log.Fatal("Pool threads is more than 0 !")
	}
	
	nnp := &NetworkPool{
		NetworkPoolOptions: opt,
		loader:             make(chan loader_resource, opt.ActionBox.LenResource()),
		wg:                 &sync.WaitGroup{},
	}
	
	nnp.goPool()
	
	for _ = range opt.ActionBox.LenResource() {
		nnp.wg.Add(1)
		nnp.loader <- loader_resource{res: opt.ActionBox.GetResource()}
	}
	
	nnp.wg.Wait()
	nnp.Close()
}
