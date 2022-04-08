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
	// ErrCouldNotLogin login failed
	ErrCouldNotLogin = errors.New("could not login")
	// ErrWrongCaptcha the captcha is wrong
	ErrWrongCaptcha = errors.New("login: wrong captcha")
)

type loginForm struct {
	Username           string `url:"userbh"`
	Password           string `url:"pas2s"`
	VCode              string `url:"vcode"`
	CW                 string `url:"cw"`
	ViewState          string `fill:"__VIEWSTATE" url:"__VIEWSTATE"`
	ViewStateGenerator string `fill:"__VIEWSTATEGENERATOR" url:"__VIEWSTATEGENERATOR"`
	ViewStateEncrypted string `fill:"__VIEWSTATEENCRYPTED" url:"__VIEWSTATEENCRYPTED"`
	XZBZ               string `url:"xzbz"` // default to 1
	YXDM               string `fill:"yxdm" url:"yxdm"`
}

// login 登录系统
func (c *punchClient) login(account *Account) (err error) {
	var req *http.Request
	loginURL := host + "/login.aspx"
	req, err = getWithContext(c.ctx, loginURL)
	if err != nil {
		return
	}
	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return
	}
	f := &loginForm{}
	err = parseForm(res.Body, f)
	drainBody(res.Body)
	if err != nil {
		return
	}
	req, err = getWithContext(c.ctx, host+"/Vcode.ASPX")
	if err != nil {
		return
	}
	// try three times
	for i := 0; i < 3; i++ {
		if res, err = c.httpClient.Do(req); err != nil {
			return
		}
		var vImg []byte
		vImg, err = io.ReadAll(res.Body)
		res.Body.Close()
		if f.VCode, err = captcha.Recognize(vImg); err != nil {
			return
		}
		if len(f.VCode) == 4 {
			break
		}
		t := time.NewTimer(time.Second)
		select {
		case <-t.C:
		case <-c.ctx.Done():
			t.Stop()
			err = c.ctx.Err()
			return
		}
	}
	if len(f.VCode) != 4 {
		return errors.New("cannot recognize vcode")
	}

	f.Username = account.Username
	hash := md5.New()
	_, err = hash.Write([]byte(strings.ToUpper(account.Password)))
	if err != nil {
		return
	}
	f.Password = hex.EncodeToString(hash.Sum(nil))
	f.XZBZ = "1"
	var value url.Values
	if value, err = query.Values(f); err != nil {
		return
	}

	req, err = postFormWithContext(c.ctx, loginURL, value)
	if err != nil {
		return
	}

	c.httpClient.CheckRedirect = notRedirect
	if res, err = c.httpClient.Do(req); err != nil {
		return
	}
	c.httpClient.CheckRedirect = nil
	defer drainBody(res.Body)

	if res.StatusCode == http.StatusFound { // redirect after login success
		return
	}
	_, cw, _ := parseHTML(bufio.NewReader(res.Body), `<input name="cw"`)

	switch cw {
	case "验证码错误!":
		err = ErrWrongCaptcha
	case "":
		err = ErrCouldNotLogin
	default:
		err = fmt.Errorf("login failed: %s", cw)
	}
	return
}

func parseForm(body io.Reader, form *loginForm) (err error) {
	bufferReader := bufio.NewReader(body)
	const inputElement = "<input type=\"hidden\""

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
