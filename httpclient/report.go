package httpclient

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

// index for resData
const (
	indexWID uint8 = iota
	indexHealthFormData
)

type htmlSymbol byte

const (
	symbolJSON htmlSymbol = iota
	symbolString
)

var (
	prefixArray = [...]string{"var _selfFormWid", "fillDetail"}
	symbolArray = [...]htmlSymbol{symbolString, symbolJSON}

	// Error
	ErrCannotParseData = errors.New("data: parse error")
	ErrInvalidSymbol   = errors.New("data: invalid symbol")
)

// getFormSessionID 获取打卡系统的SessionID
func getFormSessionID(ctx context.Context, cookies []*http.Cookie) ([]*http.Cookie, error) {
	req, err := getWithContext(ctx, "http://form.hhu.edu.cn/pdc/form/list")
	if err != nil {
		return nil, err
	}
	setCookies(req, cookies)

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	res.Body.Close()

	if cookies := getCookie(res.Cookies(), []string{"JSESSIONID"}); cookies != nil {
		return cookies, nil
	}
	return nil, CookieNotFoundErr{"JSESSIONID"}
}

// getFormDetail 获取打卡表单详细信息
func getFormDetail(ctx context.Context, cookies []*http.Cookie) (*HealthForm, *QueryParam, error) {
	req, err := getWithContext(ctx, "http://form.hhu.edu.cn/pdc/formDesignApi/S/gUTwwojq")
	if err != nil {
		return nil, nil, err
	}
	setCookies(req, cookies)

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, nil, err
	}

	var reader io.ReadCloser
	if reader, err = responseReader(res); err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	bufferReader := bufio.NewReader(reader)

	var line string
	for err == nil && line != "<script type=\"text/javascript\">" {
		line, err = scanLine(bufferReader)
	}

	var (
		resData [2][]byte // wid, healthFormData
		index   = 0
	)

	for err == nil && index != 2 {
		line, err = scanLine(bufferReader)
		if strings.HasPrefix(line, prefixArray[index]) {
			if resData[index], err = parseData(line, symbolArray[index]); err != nil {
				return nil, nil, err
			}
			index++
		}
	}

	if index != 2 {
		return nil, nil, ErrCannotParseData
	}

	form := &HealthForm{}

	if err = json.Unmarshal(resData[indexHealthFormData], form); err != nil {
		return nil, nil, err
	}

	form.DataTime = time.Now().In(timeZone).Format("2006/01/02") // 表单中增加打卡日期

	params := &QueryParam{
		Wid:    string(resData[indexWID]),
		UserID: form.StudentID,
	}

	return form, params, nil
}

// postForm 提交打卡表单
func postForm(ctx context.Context, form *HealthForm, params *QueryParam, cookies []*http.Cookie) error {
	value, err := query.Values(form)
	if err != nil {
		return err
	}

	var req *http.Request
	req, err = postWithContext(ctx,
		"http://form.hhu.edu.cn/pdc/formDesignApi/dataFormSave",
		value,
	)
	if err != nil {
		return err
	}

	setCookies(req, cookies)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	value, err = query.Values(params)
	if err != nil {
		return err
	}

	req.URL.RawQuery = value.Encode()

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("post failed, status code: %d", res.StatusCode)
	}
	return nil
}

func parseData(data string, symbol htmlSymbol) (res []byte, err error) {
	switch symbol {
	case symbolJSON:
		res, err = getSlice(data, '{', '}')
	case symbolString:
		res, err = getSlice(data, '\'', '\'')
		res = res[1 : len(res)-1]
	default:
		err = ErrInvalidSymbol
	}
	return
}

func getSlice(data string, startSymbol, endSymbol byte) (res []byte, err error) {
	start := strings.IndexByte(data, startSymbol)
	if start == -1 {
		return nil, ErrCannotParseData
	}

	length := strings.IndexByte(data[start+1:], endSymbol)
	if length == -1 {
		return nil, ErrCannotParseData
	}

	res = make([]byte, length+2)
	copy(res, data[start:])

	return res, err
}
