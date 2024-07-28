package network_pool

import (
	"github.com/GoHippo/network/fasthttp_client"
	"github.com/GoHippo/slogpretty/sl"
)

func (nnp *NetworkPool) goPool() {
	run := func() {
		for {
			
			resource := <-nnp.loader
			
			if exit, ok := resource.res.(string); ok && exit == "EXIT" {
				return
			}
			cli, err := nnp.getClient()
			
			if err != nil {
				if nnp.Log != nil {
					nnp.Log.Error("Error get http client in pool", sl.Err(err))
				}
				nnp.wg.Done()
				continue
			}
			
			nnp.ActionBox.Check(cli, resource.res)
			cli.Close()
			
			if nnp.FuncSignalDone != nil {
				nnp.FuncSignalDone(1)
			}
			
			nnp.wg.Done()
		}
	}
	
	for _ = range nnp.Threads {
		go run()
	}
}

func (nnp *NetworkPool) Close() {
	for _ = range nnp.Threads {
		nnp.loader <- loader_resource{res: "EXIT"}
	}
	
	close(nnp.loader)
}

func (nnp *NetworkPool) getClient() (*fasthttp_client.FasthttpClient, error) {
	if nnp.CliOptions.ProxyUse && nnp.CliOptions.ProxyService.GetCountProxyImap() != 0 {
		return fasthttp_client.NewFasthttpClient(nnp.CliOptions)
	}
	
	if nnp.clientStatic != nil {
		return nnp.clientStatic, nil
	}
	
	var err error
	nnp.clientStatic, err = fasthttp_client.NewFasthttpClient(nnp.CliOptions)
	return nnp.clientStatic, err
}
