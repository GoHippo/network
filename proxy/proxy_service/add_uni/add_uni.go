// Package add_uni Пакет не доработан  и не оттестирован.

package add_uni

import (
	"fmt"
	"github.com/GoHippo/network/fasthttp_client/custom_pool"
	"github.com/GoHippo/network/proxy/checker_proxy"
	"github.com/GoHippo/network/proxy/proxy_service"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// ====================== AddUniversal ======================

// AddUniOptions Указать PathFile или ArrLines. Взависимости от этого будет выбран режим работы.
type AddUniOptions struct {
	PathFile  string
	ArrLines  []string
	Ps        *proxy_service.ProxyService
	Timeout   time.Duration
	Threads   int
	FuncADD   func(i int)
	Log       *slog.Logger
	countAdd  int
	lockCount *sync.Mutex
}

func NewAddProxyUni(o AddUniOptions) (int, error) {
	var op = `proxy_service.NewAddProxyUni`
	if o.PathFile != "" {
		if err := o.readFile(); err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	if len(o.ArrLines) == 0 {
		return 0, fmt.Errorf("%s: not lines proxy for add", op)
	}

	o.lockCount = new(sync.Mutex)
	custom_pool.NewNetworkPool(custom_pool.NetworkPoolOptions{
		ActionBox:      &o,
		Threads:        o.Threads,
		Log:            o.Log,
		FuncSignalDone: o.FuncADD,
	})

	return o.countAdd, nil
}

// ====================== ActionBox ======================

func (o *AddUniOptions) Check(resource any) {
	res := resource.(string)

	if config, ok := checker_proxy.CheckProxyUni(res, o.Timeout); ok {
		o.Ps.AddProxy(config)

		o.lockCount.Lock()
		o.countAdd++
		o.lockCount.Unlock()
	}

	if o.FuncADD != nil {
		o.FuncADD(1)
	}

}
func (o *AddUniOptions) LenResource() int {
	return len(o.ArrLines)
}
func (o *AddUniOptions) GetResource() any {
	res := o.ArrLines[0]
	o.ArrLines = o.ArrLines[1:]
	return res
}

// ====================== Msg ======================

func (o *AddUniOptions) readFile() error {
	var op = `proxy_service.AddProxyUniFromFile`

	if _, err := os.Stat(o.PathFile); os.IsNotExist(err) {
		return fmt.Errorf("%v: file %s not exists", op, o.PathFile)
	}

	file, err := os.ReadFile(o.PathFile)
	if err != nil {
		return fmt.Errorf("%v: file %s not open - %w", op, o.PathFile, err)
	}

	o.parseLines(strings.Split(string(file), "\n"))

	if len(o.ArrLines) == 0 {
		return fmt.Errorf("%v: not lines proxy for add path - %v", op, o.PathFile)
	}

	return nil

}

func (o *AddUniOptions) parseLines(lines []string) {
	var op = `proxy_service.parseLines`

	for _, line := range lines {
		line = strings.TrimSpace(line)
		sp := strings.Split(line, ":")
		if len(sp) != 2 {
			o.Log.Error("invalid proxy config line", slog.String("func", op), slog.String("line", line))
			continue
		}
		o.ArrLines = append(o.ArrLines, line)
	}
}
