package proxy_jar

import "log/slog"

type JarProxy struct {
	log       *slog.Logger
	ArrConfig []ProxyConfig
}

type ProxyConfig struct {
	Addr          string
	Scheme        ProxyScheme
	Host          string
	IsImapSupport bool
	// IsActive bool
	
}

type ProxyScheme string

const (
	O_HTTPS   ProxyScheme = "https"
	O_SOCKS4  ProxyScheme = "socks4"
	O_SOCKS4a ProxyScheme = "socks4a"
	O_SOCKS5  ProxyScheme = "socks5"
)

func (fk *JarProxy) AddProxy(p ProxyConfig) {
	for _, config := range fk.ArrConfig {
		if config.Addr == p.Addr {
			return
		}
	}
	
	fk.ArrConfig = append(fk.ArrConfig, p)
	
}

func (fk *JarProxy) DeleteProxy(p ProxyConfig) {
	for i, config := range fk.ArrConfig {
		if config.Addr == p.Addr {
			fk.ArrConfig = append(fk.ArrConfig[:i], fk.ArrConfig[i+1:]...)
			
			return
		}
	}
}

func (fk *JarProxy) GetCountProxy() (count int) {
	count = len(fk.ArrConfig)
	return count
}

func (fk *JarProxy) GetCountImapProxy() (count int) {
	for _, config := range fk.ArrConfig {
		if config.IsImapSupport {
			count++
		}
	}
	return count
}
