package httpclient

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/yin1999/healthreport/v2/utils"
	"github.com/yin1999/healthreport/v2/utils/captcha"
)

var (
	// ErrWrongCaptcha the captcha is wrong
	ErrWrongCaptcha = errors.New("login: wrong captcha")
	// ErrCannotRecognizeCaptcha could not recognize captcha
	ErrCannotRecognizeCaptcha = errors.New("login: cannot recognize captcha")
	// loginFields part of fields for login form,
	// must be sorted
	loginFields = [...]string{"__VIEWSTATE", "__VIEWSTATEENCRYPTED", "__VIEWSTATEGENERATOR",
		"cw", "yxdm",
	}
)

const loginURL = host + "/login.aspx"

// login 登录系统
func (c *punchClient) login(account *Account) (err error) {
	hash := md5.New()
	_, err = hash.Write([]byte(strings.ToUpper(account.Password)))
	if err != nil {
		return
	}
	form := make(url.Values, len(loginFields)+4) // 4 is for account, password, xzbz, vcode
	form.Set("userbh", account.Username)
	form.Set("pas2s", hex.EncodeToString(hash.Sum(nil)))
	form.Set("xzbz", "1")
	err = loginGet(c, form)
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ { // 重试 3 次
		err = loginPost(c, form)
		switch err {
		case ErrWrongCaptcha, ErrCannotRecognizeCaptcha:
			if utils.Wait(c.ctx, time.Second*2) != nil {
				return
			}
		default:
			return
		}
	}
	return
}

func loginGet(c *punchClient, form url.Values) error {
	req, err := getWithContext(c.ctx, loginURL)
	if err != nil {
		return err
	}
	var res *http.Response
	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer drainBody(res.Body)
	return fillMap(res.Body, form, loginFormShouldFill)
}

func loginFormShouldFill(key string) bool {
	i := sort.SearchStrings(loginFields[:], key)
	return i < len(loginFields) && loginFields[i] == key
}

func loginPost(c *punchClient, form url.Values) (err error) {
	var vcode string
	vcode, err = recognizeCaptcha(c)
	if err != nil {
		return
	}
	form.Set("vcode", vcode)

	var req *http.Request
	req, err = postFormWithContext(c.ctx, loginURL, form)
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
	err = fillMap(res.Body, form, loginFormShouldFill) // if login failed, parse form data for next post
	if err != nil {
		return
	}
	switch v := form.Get("cw"); v { // parse error message
	case "验证码错误!":
		err = ErrWrongCaptcha
	default:
		err = fmt.Errorf("login failed: %s", v)
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
		var vImg []byte
		if vImg, err = readImage(res); err != nil {
			return
		}

		if vcode, err = captcha.Recognize(vImg); err != nil {
			return
		}
		if len(vcode) == 4 {
			return
		}
		if err = utils.Wait(c.ctx, time.Second); err != nil {
			return
		}
	}
	err = ErrCannotRecognizeCaptcha
	return
}

func readImage(res *http.Response) (data []byte, err error) {
	defer drainBody(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get captcha image failed: %s", res.Status)
	}
	var img image.Image
	img, err = jpeg.Decode(res.Body)
	if err == nil {
		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, img, nil)
		if err == nil {
			data = buf.Bytes()
		}
	}
	return
}
