package cookies_jar

import (
	"strings"
	"time"
	
	"github.com/valyala/fasthttp"
)

func (j *Jar) ParseFromResponse(uri *fasthttp.URI, resp *fasthttp.Response) []CookieType {
	host := string(uri.Host())
	
	var arr_cookies []CookieType
	resp.Header.VisitAllCookie(func(key []byte, value []byte) {
		rc := respCookies("")
		arr_cookies = append(arr_cookies, rc.parseMain(host, key, value))
	})
	// fmt.Println(arr_cookies)
	return arr_cookies
}

type respCookies string

// парсит из строк VisitAllCookie() в CookieType
func (rc respCookies) parseMain(host string, key []byte, value []byte) CookieType {
	v := strings.TrimPrefix(string(value), string(key)+"=")
	if sp := strings.Split(v, ";"); len(sp) != 0 {
		v = sp[0]
	}
	
	cookie := CookieType{
		Domain: host,
		Name:   string(key),
		Value:  strings.TrimSpace(v),
	}
	allData := rc.parseElements(value)
	
	return rc.parseAll(cookie, allData)
}

// Парсит все элементы, кроме имени и значения.
func (rc respCookies) parseElements(values []byte) map[string]string {
	
	elements := make(map[string]string)
	
	for i, el := range strings.Split(string(values), ";") {
		if i == 0 {
			continue
		}
		
		sp := strings.Split(el, "=")
		
		switch len(sp) {
		case 2:
			elements[sp[0]] = sp[1]
		case 1:
		default:
			// Исправляет ошибку с несколькими знаками "="
			sp[1] = strings.Join(sp[1:], "=")
			sp = sp[:2]
		}
		
		k := strings.TrimSpace(sp[0])
		
		var v string
		if len(sp) != 1 {
			v = strings.TrimSpace(sp[1])
		}
		elements[k] = v
		
	}
	return elements
}

// Парсит найденные элементы в CookieType.
func (rc respCookies) parseAll(cookie CookieType, data map[string]string) CookieType {
	for key, value := range data {
		switch strings.ToLower(key) {
		// case "domain":cookie.Domain=arr[1]
		case "path":
			cookie.Path = value
		case "expires":
			cookie.Expires, _ = time.Parse(time.RFC1123, strings.Replace(value, "-", " ", 2))
		case "domain":
			cookie.Domain = value
		}
	}
	return cookie
}
