package httpclient

import (
	"context"
	"net/http"
)

type punchClient struct {
	ctx        context.Context
	httpClient *http.Client
}

func (c *punchClient) retryGet(req *http.Request, retry uint8) (resp *http.Response, err error) {
	for retry != 0 {
		select {
		case <-c.ctx.Done():
			if err == nil {
				err = c.ctx.Err()
			}
			return
		default:
		}
		if resp, err = c.httpClient.Do(req); err == nil {
			break
		}
		retry--
	}
	return
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
