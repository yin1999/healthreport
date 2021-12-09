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

// getCookieByName
func (cookies cookieJar) getCookieByName(name string) (res []*http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			res = append(res, cookie)
		}
	}
	return
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

// getCookieByDomain use domain as filter to get cookies
func (cookies cookieJar) getCookieByDomain(domain string) (res []*http.Cookie) {
	for _, cookie := range cookies {
		if strings.HasSuffix(domain, cookie.Domain) {
			res = append(res, cookie)
		}
	}
	return
}

// Cookies get cookie by domains
func (j cookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.getCookieByDomain(u.Hostname())
}
