package httpclient

import (
	"net/http"
	"net/url"
	"strings"
)

type customCookieJar interface {
	http.CookieJar
	GetCookieByDomain(domain string) []*http.Cookie
	GetCookieByName(name string) []*http.Cookie
}

type cookieJar struct {
	cookies []*http.Cookie
}

// newCookieJar return a cookiejar
func newCookieJar() customCookieJar {
	return &cookieJar{}
}

// GetCookieByName
func (j *cookieJar) GetCookieByName(name string) (res []*http.Cookie) {
	for i := range j.cookies {
		if j.cookies[i].Name == name {
			res = append(res, j.cookies[i])
		}
	}
	return
}

// SetCookies set cookies to cookie storage
func (j *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for i := range cookies {
		if cookies[i].Domain == "" { // if cookie.Domain is empty, using host instead
			cookies[i].Domain = u.Hostname()
		}
		j.cookies = append(j.cookies, cookies[i])
	}
}

// GetCookieByDomain use domain as filter to get cookies
func (j *cookieJar) GetCookieByDomain(domain string) (res []*http.Cookie) {
	for i := range j.cookies {
		if strings.HasSuffix(j.cookies[i].Domain, domain) {
			res = append(res, j.cookies[i])
		}
	}
	return
}

// Cookies get cookie by domains
func (j *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.GetCookieByDomain(u.Hostname())
}
