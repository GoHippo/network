package add_uni

import (
	"github.com/GoHippo/network/proxy/proxy_service"
	"github.com/GoHippo/pterm_tools/pterm_pb"
	"github.com/GoHippo/slogpretty/slogpretty"
	"log/slog"
	"testing"
	"time"
)

func TestAddUni(t *testing.T) {
	var log = slogpretty.SetupPrettySlog(slog.LevelInfo)
	ps := proxy_service.NewProxyService(log, 5)

	pb := pterm_pb.NewPB(500, "check proxy")
	count, err := NewAddProxyUni(AddUniOptions{
		PathFile: "/home/meteoroot/expanse/project/GO/GoHippo/network/proxy/proxy_service/test.txt",
		ArrLines: nil,
		Ps:       ps,
		Timeout:  time.Second * 5,
		Threads:  100,
		FuncADD:  pb.Add,
		Log:      log,
	})

	if err != nil {
		t.Error(err)
	}
	log.Info("Add proxy", slog.Int("count", count))
}
