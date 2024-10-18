package cookies_jar

import (
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Метод конвертирует из (cookies_jar.CookieType) в string_netscape
func (j *Jar) ConvToNetscape() []byte {

	var netscapeF []byte
	for _, arrCookie := range j.jar {
		for _, cookie := range arrCookie {
			str := cookie.Domain + "	"

			httpOnly := strconv.FormatBool(cookie.HttpOnly)
			str = str + httpOnly + "	"

			str = str + cookie.Path + "	"

			secure := strconv.FormatBool(cookie.Secure)
			str = str + secure + "	"

			expire := strconv.Itoa(int(cookie.Expires.Unix()))
			str = str + expire + "	"

			str = str + cookie.Name + "	"

			str = str + cookie.Value + "\n"
			netscapeF = append(netscapeF, []byte(str)...)
		}
	}
	return netscapeF
}

// Функция конвертурует массив строк формата netscape в массив строк формата cookies_jar.CookieType
func (j *Jar) ConvNetscapeToCookieType(arr_str []string) {

	for _, str := range arr_str {
		elements := strings.Split(str, "\t")
		if len(elements) != 7 {
			continue
		}

		httpOnly, _ := strconv.ParseBool(strings.ToLower(elements[1]))
		secure, _ := strconv.ParseBool(strings.ToLower(elements[3]))

		cookie := CookieType{
			Domain:   elements[0],
			HttpOnly: httpOnly,
			Path:     elements[2],
			Secure:   secure,
			Name:     elements[5],
			Value:    elements[6],
		}

		if len(elements[4]) > 10 {
			elements[4] = elements[4][:10]
		}

		expires, err := strconv.Atoi(elements[4])
		if err != nil {
			if j.log != nil {
				j.log.Error("Ошибка в парсе строке в элементе Expires:"+elements[4], slog.String("err", err.Error()))
			}
			expires = int(time.Now().Add(time.Hour * 999).Unix())
		}
		cookie.Expires = time.Unix(int64(expires), 0)

		j.SetCookiesJar(cookie)
	}

}
