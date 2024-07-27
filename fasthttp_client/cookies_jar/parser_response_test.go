package cookies_jar

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestRespCookiesTest(t *testing.T) {
	rs := respCookies("")
	fCheckCookie := func(cookie CookieType, k, v, p string) {
		if cookie.Name != k {
			t.Error("cookie name not match!")
		}

		if cookie.Value != v {
			t.Error("cookie value not match!")
		}

		if cookie.Path != p {
			t.Error("cookie path not match!")
		}
	}

	key := `__Host-GAPS`
	value := `__Host-GAPS=1:Es8cPmd8O5c9-vrho7oCfKiGmM4k0g:ePKnTkgZu62G7iAI;Path=/;Expires=Mon, 08-Jun-2026 13:15:58 GMT;Secure;HttpOnly;Priority=HIGH`

	fCheckCookie(rs.parseMain("accounts.google.com", []byte(key), []byte(value)), key, `1:Es8cPmd8O5c9-vrho7oCfKiGmM4k0g:ePKnTkgZu62G7iAI`, "/")

	key = `XSRF-TOKEN`
	value = `XSRF-TOKEN=af443153437a4502aabe914620690f55; Path=/id`
	fCheckCookie(rs.parseMain("accounts.google.com", []byte(key), []byte(value)), key, `af443153437a4502aabe914620690f55`, "/id")

	key = `XSRF-TOKEN`
	value = `XSRF-TOKEN=; Path=/id`
	fCheckCookie(rs.parseMain("accounts.google.com", []byte(key), []byte(value)), key, ``, "/id")
}

func TestFromRespCookies(t *testing.T) {
	jar := NewJar(nil)
	req := fasthttp.AcquireRequest()
	req.SetRequestURI("https://login.live.com/")

	resp := fasthttp.AcquireResponse()

	resp.Header.Set("Set-Cookie", "__Host-MSAAUTHP=11-M.C521_BL2.0.U.CnCMP0zmbVGS3HiJf!HenKe!Sv0RiMmY95UDm6AXJ2h!jaQSiAunIi38zwaIP0*WlNTWKyN9OstLsAro9V8mF34RNNy8cfJBlII2TreJRFN4FvBYTKbkE0fQQKMuvRLzO7R3rI70TOMjFIoD*Zl85v0JVCs1H!nCZhhrnT**Gf0JM4WkLwSH6PzqfX3o0cJgJZZ685mPpXeCZYyi1BPHrxv2GA5L5CrOFWZfyAKil8fBiyMU31BE4bof5Ek9xgK4jKSOfH6XM9kVUtynWfvz3IbIi2U5E16xZq3!nYVxVwpmpTAlSe*EtNAE5ukzRK38ZWTjCLelOOyvUdFdCS*5KSmAD2ensjYAMb8v8DIVAEtSIqReQWukYchI6T4nOIWpWancCcCGYLd6hSr3ke17g0XYJ1EytuZ9B*!pmgZYM4tXguaZb!JA5xJFbgJWNnOFg897uVvbiDif9Kie6Tdvf9F8ufUz3Lq*oesrMFP7LexHBqzicf9i0BizIx9ACfDIXTIYpTiBbdmAxze0YW0DAv8ns4uvIDBIN*0MuLIJkVyhOov5VNSUG2K2G3AlEDhEYzReoffbZyxgdcPdH!hW*jhJcy1T4KCAyLoJkMnvLL1DUdiQDVlBaqwhh*5soCrTrTyCcyBWM9Z7RlhXJls2vkm6yzYCfI7NohMimlckyGr1; expires=Sat, 05-Jul-2025 19:28:25 GMT; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "uaid=a3a00687286b49a8948503d9966a84ce; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "__Host-MSAAUTH=11; expires=Thu, 30-Oct-1980 16:00:00 GMT; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSCC=109.248.148.53-LV; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPRequ=id=N&lt=1718047705&co=1; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "__Host-MSAAUTHP=11-M.C521_BL2.0.U.CnCMP0zmbVGS3HiJf!HenKe!Sv0RiMmY95UDm6AXJ2h!jaQSiAunIi38zwaIP0*WlNTWKyN9OstLsAro9V8mF34RNNy8cfJBlII2TreJRFN4FvBYTKbkE0fQQKMuvRLzO7R3rI70TOMjFIoD*Zl85v0JVCs1H!nCZhhrnT**Gf0JM4WkLwSH6PzqfX3o0cJgJZZ685mPpXeCZYyi1BPHrxv2GA5L5CrOFWZfyAKil8fBiyMU31BE4bof5Ek9xgK4jKSOfH6XM9kVUtynWfvz3IbIi2U5E16xZq3!nYVxVwpmpTAlSe*EtNAE5ukzRK38ZWTjCLelOOyvUdFdCS*5KSmAD2ensjYAMb8v8DIVAEtSIqReQWukYchI6T4nOIWpWancCcCGYLd6hSr3ke17g0XYJ1EytuZ9B*!pmgZYM4tXguaZb!JA5xJFbgJWNnOFg897uVvbiDif9Kie6Tdvf9F8ufUz3Lq*oesrMFP7LexHBqzicf9i0BizIx9ACfDIXTIYpTiBbdmAxze0YW0DAv8ns4uvIDBIN*0MuLIJkVyhOov5VNSUG2K2G3AlEDhEYzReoffbZyxgdcPdH!hW*jhJcy1T4KCAyLoJkMnvLL1DUdiQDVlBaqwhh*5soCrTrTyCcyBWM9Z7RlhXJls2vkm6yzYCfI7NohMimlckyGr1; expires=Sat, 05-Jul-2025 19:28:25 GMT; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "PPLState=1; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=.live.com; Secure; path=/; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPShared= ; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPPre=revee-linux%40outlook.com%7c7866589a44a6723f%7c%7c; domain=login.live.com; path=/; Expires=Sat, 05-Jul-2025 19:28:25 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPCID=7866589a44a6723f; HTTPOnly; domain=login.live.com; path=/; Expires=Sat, 05-Jul-2025 19:28:25 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPAuth=Disabled; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPProf=Disabled; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "NAP=V=1.9&E=1d84&C=-S3Xjr1bmiOUeSzPqL3TDxteKNFhHaOI6KwG15u0P8tmSMmDiBWbTw&W=1; HTTPOnly; domain=.live.com; path=/; Expires=Thu, 19-Sep-2024 02:28:25 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "ANON=A=F9DDA2A0D0E9C11137AE65D5FFFFFFFF&E=1dde&W=1; HTTPOnly; domain=.live.com; path=/; Expires=Sat, 28-Dec-2024 03:28:25 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPVis=; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPVisNet= ; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPNSVis=; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPNSVisNet=; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None")
	resp.Header.Set("Set-Cookie", "WLSSC=EgAsAgMAAAAMgAAAmAABlyV0gXwancDLjsTWHb64TUGfXJC96/tD4isNrV3dsk+eQ488fS67a26Q1l/tfWOU7co7/dZRpp6Nzq6j+YNajPzuk6nf7137IUiepKlAbyZghwu1ZhppI1L/wRZ0WAbtcmvUrrcByXXMQzBDnM3yLnyk2qyHN2MPXnUoU+OT/EOOVqjN9TLjdezPxri5YqtItAizPivHH1YjWrKroj2GvVF+AhpF6CepbM/VVPiWADrdvpftD3wfhdxGU1cSfOW21cyKZXq5N8SYcJoVDP7z0P2At7Mxa5HnL19ksv6DXLNG7zXKAJsfZbiRLysCk9Nf9avQIF/n+bHQnMEg14n5eBsBfgAbAf2/AwD1mYlb2VNnZpfp+2UQJwAAChOgABAYAHJldmVlLWxpbnV4QG91dGxvb2suY29tAFwAACRyZXZlZS1saW51eCVvdXRsb29rLmNvbUBwYXNzcG9ydC5jb20AAAH8UlUAAAAAAAAEGQIAAIcWVUAABkMABUxpbnV4AAVSYXZlZQAAAAAAAAAAAAAAAAAAAAAAAESmcj94ZliaAADZU2dml5ByZgAAAAAAAAAAAAAAAA8AMTA5LjI0OC4xNDguNTMABAEAAAAAAAAAAAAAAAAQBAAAAAAAAAAAAAAAAAAAAJt/LrC3CrftAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAwAAAA==; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=.live.com; Secure; path=/; SameSite=None; HttpOnly\n\t\t")
	resp.Header.Set("Set-Cookie", "MSPAPPVis=; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "MSPAPPVisNet=; expires=Thu, 30-Oct-1980 16:00:00 GMT; domain=login.live.com; Secure; path=/; SameSite=None")
	resp.Header.Set("Set-Cookie", "SDIDC=CsJNeXAHEzzHohjr1!Ufn1Pag8MijYjs!PSnAuzLfs9WLMIERkFU8DKDC2el0H09rfGO1fbgejfPxsGZcXbKweNGXggWXLRSgtSAv!IxyDO2k7tLELUrA4945dd8n9ESFVs3E0GrIOpDqNj!dXfLI00$; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "JSHP=3$revee-linux%40outlook.com$Linux$Ravee$$2$0$0$12695224707288861367$0; domain=login.live.com; path=/; Expires=Sat, 05-Jul-2025 19:28:25 GMT; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "JSH=3$revee-linux%40outlook.com$Linux$Ravee$$2$0$0$12695224707288861367$0; domain=login.live.com; path=/; Secure; SameSite=None")
	resp.Header.Set("Set-Cookie", "MSPSoftVis=@72198325083833620@:@; expires=Sat, 05-Jul-2025 19:28:25 GMT; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")
	resp.Header.Set("Set-Cookie", "OParams=11O.DgC5WibBDxV6XINj4KryS8zbDmM5T2dnn4kduOq7hOVrIhJsQDxDBi!TVMCHHdACqmJcCDECFpT6See5BR4JOsKSndY98IetaAWMSpkTUzdukTrfu14ftjKSF45DnWxImA$$; domain=login.live.com; Secure; path=/; SameSite=None; HttpOnly")

	jar.AddFromResponse(req.URI(), resp)

	// spew.Dump(jar.jar)

	req.SetRequestURI("https://outlook.live.com/owa/0/")
	jar.SetCookiesReq(req)

	// spew.Dump(req.Header)
}
