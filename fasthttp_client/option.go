package fasthttp_client

import "github.com/GoHippo/network/fasthttp_client/cookies_jar"

type errNetworkCounter interface {
	AddCountNetworkErr(err error)
}

type DoOption struct {
	ID                  string
	DoCountReconnecting int
	NotBodyDecode       bool
	Jar                 *cookies_jar.Jar
	ErrCounter          errNetworkCounter
}

/*func (fco FastHttpClientOptions) GetDoOption(id string) DoOption {
	return DoOption{
		ID:                  id,
		DoCountReconnecting: fco.CountReconnections,
		NotBodyDecode:       false,
		Jar:                 fco.Jar,
	}
}*/
