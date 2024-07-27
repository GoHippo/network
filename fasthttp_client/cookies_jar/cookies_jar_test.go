package cookies_jar

import (
	"log"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

// TODO: когда нибудь написать тесты для Jar
func TestCookiesJar(t *testing.T) {
	jar := NewJar(nil)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	client := &fasthttp.Client{
		Dial: fasthttpproxy.FasthttpHTTPDialer("127.0.0.1:6666"),
	}

	req.SetRequestURI(`https://accounts.google.com`)

	if err := client.Do(req, resp); err != nil {
		log.Fatal(err)
	}

	host := string(req.URI().Host())

	jar.AddFromResponse(req.URI(), resp)
	if _, ok := jar.jar[".accounts.google.com"]; !ok {
		t.Error("AddFromResponse fail")
	}

	testCookie := CookieType{
		Domain:  "accounts.google.com",
		Path:    "/",
		Expires: time.Now(),
		Name:    "__Host-GAPS",
		Value:   "koki",
	}

	jar.SetCookiesJar(testCookie)
	if c := jar.GetByNameCookie(host, "__Host-GAPS", false); c.Value != "koki" {
		t.Error("SetCookiesJar fail")
	}

	if c := jar.GetByNameCookie(host, "__Host-GAPS", false); c.Name != "__Host-GAPS" {
		t.Error("GetByNameCookie fail")
	}

	jar.SetCookiesReq(req)
	if string(req.Header.Cookie(`__Host-GAPS`)) == "" {
		t.Error("SetCookiesReq fail")
	}

	jar.DeleteDomain("accounts.google.com", false)
	if _, ok := jar.jar[".accounts.google.com"]; ok {
		t.Error("DeleteDomain fail")
	}

	if c := jar.GetCookieFromResponse(req.URI(), resp, "__Host-GAPS", false); c.Name != "__Host-GAPS" {
		t.Error("GetCookieFromResponse fail")
	}

}

func TestSetCookie(t *testing.T) {
	jar := NewJar(nil)

	/*req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	client := &fasthttp.Client{
		Dial: fasthttpproxy.FasthttpHTTPDialer("127.0.0.1:6666"),
	}*/

	cook := CookieType{
		Domain:  "api.test.com",
		Path:    "/",
		Expires: time.Now(),
		Name:    "NameTest",
		Value:   "ValueTest",
	}

	jar.SetCookiesJar(cook)
	if len(jar.jar) == 0 {
		t.Error("[0] SetCookiesJar fail!")
	}

	cook.Value = "SetValue"
	jar.SetCookiesJar(cook)
	if len(jar.jar[".api.test.com"]) > 1 {
		t.Error("[1] SetCookiesJar fail!")
	}

	cook2 := cook
	cook2.Name = "NameTest2"
	jar.SetCookiesJar(cook2)

	cook.Value = ""
	jar.SetCookiesJar(cook)
	if len(jar.jar[".api.test.com"]) != 1 {
		t.Error("[2] SetCookiesJar fail!")
	}

	spew.Dump(jar.jar)

}

func TestSetCookiesFromResponse(t *testing.T) {
	jar := NewJar(nil)

	req := fasthttp.AcquireRequest()

	req.SetRequestURI("https://login.live.com")

	resp := fasthttp.AcquireResponse()
	resp.Header.Set("Set-Cookie", "NAP=V=1.9&E=1d84&C=ito_jjgxQQ6sn5bPMgIIqWqKQT3I96vfsjWojPjBAtXMvtn2uDjkMw&W=1; HTTPOnly; domain=.live.com; path=/; Expires=Wed, 18-Sep-2024 17:30:14 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPAuth=Disabled; expires=Sat, 05-Jul-2025 10:30:14 GMT; Secure; path=/; SameSite=None; HttpOnly")

	jar.AddFromResponse(req.URI(), resp)

	if len(jar.jar) == 0 {
		t.Error("[0] AddFromResponse fail")
	}

	spew.Dump(jar.jar)

}

func TestDeleteDomain(t *testing.T) {
	jar := NewJar(nil)

	req := fasthttp.AcquireRequest()

	req.SetRequestURI("https://login.live.com")

	resp := fasthttp.AcquireResponse()
	resp.Header.Set("Set-Cookie", "NAP=V=1.9&E=1d84&C=ito_jjgxQQ6sn5bPMgIIqWqKQT3I96vfsjWojPjBAtXMvtn2uDjkMw&W=1; HTTPOnly; domain=.live.com; path=/; Expires=Wed, 18-Sep-2024 17:30:14 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPAuth=Disabled; expires=Sat, 05-Jul-2025 10:30:14 GMT; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPAuth=Disabled; Domain=outlook.live.com; expires=Sat, 05-Jul-2025 10:30:14 GMT; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPAuth=Disabled; Domain=.live.com;expires=Sat, 05-Jul-2025 10:30:14 GMT; Secure; path=/; SameSite=None; HttpOnly")

	jar.AddFromResponse(req.URI(), resp)

	if len(jar.jar) == 0 {
		t.Error("[0] DeleteDomain fail")
	}
	spew.Dump(jar.jar)

	jar.DeleteDomain(`live.com`, true)

	spew.Dump(jar.jar)

}
