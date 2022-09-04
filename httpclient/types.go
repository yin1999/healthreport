package httpclient

import (
	"context"
	"net/http"
)

type punchClient struct {
	ctx        context.Context
	httpClient *http.Client
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
