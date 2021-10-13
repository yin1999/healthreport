package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// timeZone is used for set DataTime in HealthForm
var timeZone = time.FixedZone("CST", 8*3600)

// LoginConfirm 验证账号密码
func LoginConfirm(ctx context.Context, account interface{}, timeout time.Duration) error {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	c := newClient(ctx)
	err := c.login(account.(*Account))
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

	var path string
	path, err = c.getFormSessionID() // 获取打卡系统的cookie
	if err != nil {
		return
	}

	var (
		form   map[string]string
		params *QueryParam
	)
	form, params, err = c.getFormDetail(path) // 获取打卡列表信息
	if err != nil {
		return
	}

	err = c.postForm(form, params) // 提交表单
	return
}

func newClient(ctx context.Context) *punchClient {
	jar := newCookieJar()
	return &punchClient{
		ctx:        ctx,
		jar:        jar,
		httpClient: &http.Client{Jar: jar},
	}
}

// parseURLError 解析URL错误
func parseURLError(err error) error {
	if v, ok := err.(*url.Error); ok {
		err = v.Err
	}
	return err
}
