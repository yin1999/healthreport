package httpclient

import (
	"net/http"
	"net/url"
	"strings"
)

type cookieJar []*http.Cookie

var _ http.CookieJar = &cookieJar{} // implement http.CookieJar

// newCookieJar return a cookiejar
func newCookieJar() *cookieJar {
	return &cookieJar{}
}

// SetCookies set cookies to cookie storage
func (cookies *cookieJar) SetCookies(u *url.URL, newCookies []*http.Cookie) {
	for _, cookie := range newCookies {
		if cookie.Domain == "" { // if cookie.Domain is empty, using host instead
			cookie.Domain = u.Hostname()
		}
		*cookies = append(*cookies, cookie)
	}
}

// Cookies get cookie by domains
func (cookies cookieJar) Cookies(u *url.URL) (res []*http.Cookie) {
	domain := u.Hostname()
	for _, cookie := range cookies {
		if strings.HasSuffix(domain, cookie.Domain) {
			res = append(res, cookie)
		}
	}
	return
}
