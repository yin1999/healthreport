package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// LoginConfirm 验证账号密码
func LoginConfirm(ctx context.Context, account interface{}) error {
	c := newClient(ctx)
	err := c.login(account.(*Account))
	return parseURLError(err)
}

// Punch 打卡
func Punch(ctx context.Context, account interface{}) (err error) {
	defer func() {
		err = parseURLError(err)
	}()

	c := newClient(ctx)
	for i := 0; i < 3; i++ { // 重试 3 次
		err = c.login(account.(*Account)) // 登录，获取cookie
		switch err {
		case ErrWrongCaptcha:
			if wait(ctx, time.Second*2) != nil {
				return
			}
		default:
			break
		}
	}
	if err != nil {
		return
	}

	var form map[string]string
	form, err = c.getFormDetail() // 获取打卡列表信息
	if err != nil {
		return
	}

	err = c.postForm(form) // 提交表单
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
