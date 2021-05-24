package httpclient

import (
	"context"
	"net/url"
	"time"
)

// timeZone is used for set DataTime in HealthForm
var timeZone = time.FixedZone("CST", 8*3600)

// LoginConfirm 验证账号密码
func LoginConfirm(ctx context.Context, account [2]string, timeout time.Duration) error {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	_, err := login(ctx, account)
	cc()
	return parseURLError(err)
}

// Punch 打卡
func Punch(ctx context.Context, account [2]string, timeout time.Duration) (err error) {
	var cc context.CancelFunc
	ctx, cc = context.WithTimeout(ctx, timeout)
	defer cc()

	defer func() {
		err = parseURLError(err)
	}()

	var jar customCookieJar
	jar, err = login(ctx, account) // 登录，获取cookie
	if err != nil {
		return
	}

	err = getFormSessionID(ctx, jar) // 获取打卡系统的cookie
	if err != nil {
		return
	}

	var (
		form   *HealthForm
		params *QueryParam
	)
	form, params, err = getFormDetail(ctx, jar) // 获取打卡列表信息
	if err != nil {
		return err
	}

	err = postForm(ctx, form, params, jar) // 提交表单
	return
}

// parseURLError 解析URL错误
func parseURLError(err error) error {
	if v, ok := err.(*url.Error); ok {
		err = v.Err
	}
	return err
}
