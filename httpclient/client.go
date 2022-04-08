package httpclient

import (
	"context"
	"fmt"
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

	var schoolTerm, grade string
	schoolTerm, grade, err = c.getFormSessionID() // 获取打卡系统的cookie
	if err != nil {
		return
	}

	uri := fmt.Sprintf("/txxm/rsbulid/r_3_3_st_jkdk.aspx?xq=%s&nd=%s&msie=1", schoolTerm, grade)
	var (
		form map[string]string
	)
	form, err = c.getFormDetail(uri) // 获取打卡列表信息
	if err != nil {
		return
	}

	err = c.postForm(form, uri) // 提交表单
	return
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
