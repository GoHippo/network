package proxy_service

import (
	"fmt"
	"github.com/GoHippo/network/proxy/proxy_jar"
	"log/slog"
)

type ProxyService struct {
	log    *slog.Logger
	jar    *proxy_jar.JarProxy
	loader chan poolloader
}

type poolloader struct {
	cmd   cmd_poolloader
	proxy proxy_jar.ProxyConfig
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
)

var ERR_JAR_PROXY_NULL = fmt.Errorf("Ошибка: Proxy list пустой")
