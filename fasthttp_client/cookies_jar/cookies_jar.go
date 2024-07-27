package cookies_jar

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// структура одного кука
type CookieType struct {
	Domain   string
	HttpOnly bool
	Path     string
	Secure   bool
	Expires  time.Time
	Name     string
	Value    string
}

// банка мап(домена) с массивом куков
type Jar struct {
	jar  map[string][]CookieType
	log  *slog.Logger
	lock sync.Mutex
}

func NewJar(log *slog.Logger) *Jar {
	return &Jar{
		jar:  map[string][]CookieType{},
		log:  log,
		lock: sync.Mutex{},
	}
}

// Метод добавляет кук в банку или заменит на новый.
// Ищет, как с "." перед доменом, так и без.
// Если Value пустой, то удалит кук.
func (j *Jar) SetCookiesJar(cookie CookieType) {
	j.lock.Lock()
	defer j.lock.Unlock()

	// cookie.Value = strings.TrimSpace(cookie.Value)

	fSearch := func(c CookieType) (bool, bool, int) {
		if arr, ok := j.jar[cookie.Domain]; ok {
			for i, cookieStorage := range arr {
				if strings.ToLower(c.Name) == strings.ToLower(cookieStorage.Name) && strings.ToLower(c.Path) == strings.ToLower(cookieStorage.Path) {
					return true, true, i
				}
			}
			return true, false, 0
		}
		return false, false, 0
	}

	cookie.Domain = strings.TrimPrefix(cookie.Domain, ".")
	dom, cookieName, cookieIndex := fSearch(cookie)

	if !dom {
		cookie.Domain = "." + cookie.Domain
		dom, cookieName, cookieIndex = fSearch(cookie)
	}

	switch {
	case dom && !cookieName:
		j.jar[cookie.Domain] = append(j.jar[cookie.Domain], cookie)

	case cookieName:
		if cookie.Value == "" {
			j.jar[cookie.Domain] = append(j.jar[cookie.Domain][:cookieIndex], j.jar[cookie.Domain][cookieIndex+1:]...)
			return
		}

		j.jar[cookie.Domain][cookieIndex] = cookie

	default:
		j.jar[cookie.Domain] = []CookieType{cookie}
	}

}

func (j *Jar) DeleteDomain(domain string, contains bool) {
	j.lock.Lock()
	defer j.lock.Unlock()

	fCheckDomDelete := func(d string) {

		for domStorage, _ := range j.jar {
			if domStorage == d {
				delete(j.jar, d)
				continue
			}

			if contains && strings.Contains(domStorage, d) {
				delete(j.jar, domStorage)
			}
		}
	}

	domain = strings.TrimPrefix(domain, ".")
	fCheckDomDelete(domain)

	domain = "." + domain
	fCheckDomDelete(domain)

}

// SetCookiesReq Метод добавляет куки из банки(Jar) в запрос(*fasthttp.Request).
// Можно указать слайс имен определенных кук.
func (j *Jar) SetCookiesReq(req *fasthttp.Request, onlyCookiesName ...string) {

	uri := req.URI()
	domain := string(uri.Host())

	fSearchDomainCookies := func(d string) (arrCookies []CookieType, ok bool) {
		d = "." + d

		for dStorage, arrCookiesDomain := range j.jar {
			if strings.Contains(d, dStorage) {
				// Возможно потом добавить проверку совпадения имен в разных доменах
				arrCookies = append(arrCookies, arrCookiesDomain...)
			}
		}
		return arrCookies, len(arrCookies) > 0
	}

	fOnlyCookiesName := func(c CookieType) {
		for _, name := range onlyCookiesName {
			if strings.ToLower(name) == strings.ToLower(c.Name) {
				req.Header.SetCookie(c.Name, c.Value)
			}
		}
	}

	if arr, ok := fSearchDomainCookies(domain); ok {
		for _, c := range arr {
			if strings.Contains(string(uri.Path()), c.Path) {
				if len(onlyCookiesName) != 0 {
					fOnlyCookiesName(c)
				} else {
					req.Header.SetCookie(c.Name, c.Value)
				}
			}
		}
	}

}

// AddFromResponse Метод добавляет куки с ответа запроса(resp) в банку
func (j *Jar) AddFromResponse(uri *fasthttp.URI, resp *fasthttp.Response) {
	arrCookies := j.ParseFromResponse(uri, resp)

	if len(arrCookies) != 0 {
		for _, c := range arrCookies {
			j.SetCookiesJar(c)
		}
	}
}

// GetCookieFromResponse Метод получает кук (CookieType) по имени из ответа запроса(resp)
// Кук в ответе может быть с доменом или без.
func (j *Jar) GetCookieFromResponse(uri *fasthttp.URI, resp *fasthttp.Response, name_cookie string, contains bool) CookieType {
	arrCookies := j.ParseFromResponse(uri, resp)

	name_cookie = strings.ToLower(name_cookie)

	for _, c := range arrCookies {
		nameStorage := strings.ToLower(c.Name)

		if nameStorage == name_cookie {
			return c
		}

		if contains && strings.Contains(nameStorage, name_cookie) {
			return c
		}
	}
	return CookieType{}
}

// GetByNameCookie Функция получает кук из банки по содержании(contains) именни domain(host) и по точному именни ключа(именни) кука
// Можно послать пустой домен, тогда найдет ближайшее совпадение ключа.
func (j *Jar) GetByNameCookie(domain string, name string, contains bool) CookieType {

	name = strings.ToLower(name)

	fCheckName := func(storageName, findName string) bool {
		storageName = strings.ToLower(storageName)
		findName = strings.ToLower(findName)

		if storageName == findName {
			return true
		}

		if contains {
			return strings.Contains(storageName, findName)
		}

		return false
	}

	if domain == "" {
		for _, d := range j.jar {
			for _, cookie := range d {
				if fCheckName(cookie.Name, name) {
					return cookie
				}
			}
		}
		return CookieType{}
	}

	domain = strings.TrimPrefix(domain, ".")
	if arr, ok := j.jar[domain]; ok {
		for _, c := range arr {
			if fCheckName(c.Name, name) {
				return c
			}
		}
	}

	domain = "." + domain
	if arr, ok := j.jar[domain]; ok {
		for _, c := range arr {
			if fCheckName(c.Name, name) {
				return c
			}
		}
	}
	return CookieType{}
}
