package httpclient

import (
	"context"
	"net/http"
)

// QueryParam query param struct
type QueryParam struct {
	Wid    string `url:"wid"`
	UserID string `url:"userId"`
}

// CookieNotFoundErr error interface for Cookies
type CookieNotFoundErr struct {
	cookie string
}

func (t CookieNotFoundErr) Error() string {
	return "http: can't find cookie: " + t.cookie
}

type punchClient struct {
	ctx        context.Context
	httpClient *http.Client
	jar        *cookieJar
}

// Account account info for login
type Account struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Name get the name of the account
func (a Account) Name() string {
	return a.Username
}
