package config

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
