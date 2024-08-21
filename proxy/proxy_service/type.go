package proxy_service

import (
	"fmt"
	"github.com/GoHippo/network/proxy/proxy_service/config"
	"log/slog"
)

type ProxyService struct {
	log        *slog.Logger
	jar        map[config.ProxyConfig]int
	rate_limit int
	loader     chan poolloader
}

type poolloader struct {
	cmd   cmd_poolloader
	proxy config.ProxyConfig
	resp  chan poolloader
}

type cmd_poolloader string

const (
	ERR_NULL  = "ERR_NULL"
	WAIT_FREE = "WAIT_FREE"

	CLOSE    cmd_poolloader = "CLOSE"
	ADD      cmd_poolloader = "ADD"
	DELETE   cmd_poolloader = "DELETE"
	GET      cmd_poolloader = "GET"
	GET_IMAP cmd_poolloader = "GET_IMAP"
	FREE     cmd_poolloader = "FREE"
	COUNT    cmd_poolloader = "COUNT"
)

var ERR_JAR_PROXY_NULL = fmt.Errorf("Ошибка: Proxy list пустой")
