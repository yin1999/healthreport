package httpclient

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

// LoginConfirm 验证账号密码
func LoginConfirm(ctx context.Context, account interface{}, timeout time.Duration) error {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	c := newClient(ctx)
	err := c.login(account.(*Account))
	if err == nil {
		c.logout()
	}
	cc()
	return parseURLError(err)
}

// Punch 打卡
func Punch(ctx context.Context, account interface{}, timeout time.Duration) (err error) {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	defer cc()

	defer func() {
		err = parseURLError(err)
	}()

	c := newClient(ctx)
	err = c.login(account.(*Account)) // 登录，获取cookie
	if err != nil {
		return
	}
	defer c.logout()

	err = c.getFormSessionID(3) // 获取打卡系统的cookie

	var (
		form  map[string]formValue
		query string
	)
	form, query, err = c.getFormDetail(3) // 获取打卡列表信息
	if err != nil {
		return
	}

	err = c.postForm(form, query) // 提交表单
	return
}

// SetSslVerify when set false, insecure connection will be allowed
func SetSslVerify(verify bool) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: !verify}
}

func newClient(ctx context.Context) *punchClient {
	return &punchClient{
		ctx: ctx,
		httpClient: &http.Client{
			Jar:     newCookieJar(),
			Timeout: time.Duration(10 * time.Second),
		},
	}
}

// parseURLError 解析URL错误
func parseURLError(err error) error {
	if v, ok := err.(*url.Error); ok {
		err = v.Err
	}
	return err
}
