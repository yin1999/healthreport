package httpclient

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/yin1999/healthreport/utils/captcha"
)

var (
	// ErrWrongCaptcha the captcha is wrong
	ErrWrongCaptcha = errors.New("login: wrong captcha")
	// ErrCannotRecognizeCaptcha could not recognize captcha
	ErrCannotRecognizeCaptcha = errors.New("login: cannot recognize captcha")
)

const loginURL = host + "/login.aspx"

type loginForm struct {
	Username           string `url:"userbh"`
	Password           string `url:"pas2s"`
	VCode              string `url:"vcode"`
	CW                 string `fill:"cw" url:"cw"`
	ViewState          string `fill:"__VIEWSTATE" url:"__VIEWSTATE"`
	ViewStateGenerator string `fill:"__VIEWSTATEGENERATOR" url:"__VIEWSTATEGENERATOR"`
	ViewStateEncrypted string `fill:"__VIEWSTATEENCRYPTED" url:"__VIEWSTATEENCRYPTED"`
	XZBZ               string `url:"xzbz"` // default to 1
	YXDM               string `fill:"yxdm" url:"yxdm"`
}

// login 登录系统
func (c *punchClient) login(account *Account) (err error) {
	hash := md5.New()
	_, err = hash.Write([]byte(strings.ToUpper(account.Password)))
	if err != nil {
		return
	}
	form := &loginForm{
		Username: account.Username,
		Password: hex.EncodeToString(hash.Sum(nil)),
		XZBZ:     "1",
	}
	err = loginGet(c, form)
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ { // 重试 3 次
		err = loginPost(c, form)
		switch err {
		case ErrWrongCaptcha, ErrCannotRecognizeCaptcha:
			if wait(c.ctx, time.Second*2) != nil {
				return
			}
		default:
			return
		}
	}
	return
}

func loginGet(c *punchClient, form *loginForm) error {
	req, err := getWithContext(c.ctx, loginURL)
	if err != nil {
		return err
	}
	var res *http.Response
	res, err = c.httpClient.Do(req)
	defer drainBody(res.Body)
	return parseForm(res.Body, form)
}

func loginPost(c *punchClient, form *loginForm) (err error) {
	form.VCode, err = recognizeCaptcha(c)
	if err != nil {
		return
	}

	var value url.Values
	if value, err = query.Values(form); err != nil {
		return
	}

	var req *http.Request
	req, err = postFormWithContext(c.ctx, loginURL, value)
	if err != nil {
		return
	}

	c.httpClient.CheckRedirect = notRedirect
	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return
	}
	c.httpClient.CheckRedirect = nil
	defer drainBody(res.Body)

	if res.StatusCode == http.StatusFound { // redirect after login success
		return
	}
	err = parseForm(res.Body, form) // if login failed, parse form data for next post
	if err != nil {
		return
	}
	switch form.CW { // parse error message
	case "验证码错误!":
		err = ErrWrongCaptcha
	default:
		err = fmt.Errorf("login failed: %s", form.CW)
	}
	return
}

func recognizeCaptcha(c *punchClient) (vcode string, err error) {
	var req *http.Request
	req, err = getWithContext(c.ctx, host+"/Vcode.ASPX")
	if err != nil {
		return
	}
	// try three times
	for i := 0; i < 3; i++ {
		var res *http.Response
		if res, err = c.httpClient.Do(req); err != nil {
			return
		}
		vImg := make([]byte, res.ContentLength)
		var n int
		n, err = res.Body.Read(vImg)
		res.Body.Close()
		if n != int(res.ContentLength) && err != nil {
			err = fmt.Errorf("get captcha image failed: %w", err)
			return
		}
		if vcode, err = captcha.Recognize(vImg[:n]); err != nil {
			return
		}
		if len(vcode) == 4 {
			return
		}
		if err = wait(c.ctx, time.Second); err != nil {
			return
		}
	}
	err = ErrCannotRecognizeCaptcha
	return
}

func parseForm(body io.Reader, form *loginForm) (err error) {
	bufferReader := bufio.NewReader(body)
	const inputElement = "<input "

	var filler *structFiller
	if filler, err = newFiller(form, "fill"); err != nil {
		return
	}
	var key, value string
	for {
		key, value, err = parseHTML(bufferReader, inputElement)
		if err != nil {
			break
		}
		filler.fill(key, value)
	}
	if err == io.EOF {
		err = nil
	}
	return
}
