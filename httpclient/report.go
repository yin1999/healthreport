package httpclient

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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

type htmlSymbol uint8

const (
	symbolJSON htmlSymbol = iota
	symbolString
)

var (
	prefixArray = [...]string{"var _selfFormWid", "fillDetail"}
	symbolArray = [...]htmlSymbol{symbolString, symbolJSON}
	//ErrCannotParseData cannot parse html data error
	ErrCannotParseData = errors.New("data: parse error")
)

// getFormSessionID 获取打卡系统的SessionID
func getFormSessionID(ctx context.Context, jar customCookieJar) error {
	req, err := getWithContext(ctx, "http://dailyreport.hhu.edu.cn/pdc/form/list")
	if err != nil {
		return err
	}

	client := &http.Client{Jar: jar}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return err
	}
	res.Body.Close()

	if jar.GetCookieByDomain("dailyreport.hhu.edu.cn") == nil {
		err = CookieNotFoundErr{"JSESSIONID"}
	}
	return err
}

// getFormDetail 获取打卡表单详细信息
func getFormDetail(ctx context.Context, jar http.CookieJar) (form *HealthForm, params *QueryParam, err error) {
	var req *http.Request
	req, err = getWithContext(ctx, "http://dailyreport.hhu.edu.cn/pdc/formDesignApi/S/gUTwwojq")
	if err != nil {
		return
	}

	client := &http.Client{Jar: jar}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return
	}

	var reader io.ReadCloser
	if reader, err = responseReader(res); err != nil {
		return
	}

	var (
		bufferReader = bufio.NewReader(reader)
		resData      [2][]byte // wid, healthFormData
		index        = 0
		line         string
	)

	for err == nil && index != 2 {
		line, err = scanLine(bufferReader)
		if strings.HasPrefix(line, prefixArray[index]) {
			resData[index], err = parseData(line, symbolArray[index])
			index++
		}
	}
	reader.Close()

	if err != nil || index != 2 {
		err = ErrCannotParseData
		return
	}

	form = &HealthForm{}

	if err = json.Unmarshal(resData[indexHealthFormData], form); err != nil {
		return
	}

	form.DataTime = time.Now().In(timeZone).Format("2006/01/02") // 表单中增加打卡日期

	params = &QueryParam{
		Wid:    string(resData[indexWID]),
		UserID: form.StudentID,
	}
	return
}

// postForm 提交打卡表单
func postForm(ctx context.Context, form *HealthForm, params *QueryParam, jar http.CookieJar) error {
	value, err := query.Values(form)
	if err != nil {
		return err
	}

	var req *http.Request
	req, err = postFormWithContext(ctx,
		"http://dailyreport.hhu.edu.cn/pdc/formDesignApi/dataFormSave",
		value,
	)
	if err != nil {
		return err
	}

	value, err = query.Values(params)
	if err != nil {
		return err
	}

	req.URL.RawQuery = value.Encode()

	client := &http.Client{
		Jar: jar,
	}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("post failed, status: " + res.Status)
	}
	return nil
}

func parseData(data string, symbol htmlSymbol) (res []byte, err error) {
	switch symbol {
	case symbolJSON:
		res, err = getSlice(data, '{', '}', true)
	case symbolString:
		res, err = getSlice(data, '\'', '\'', false)
	default:
		err = errors.New("data: invalid symbol")
	}
	return
}

func getSlice(data string, startSymbol, endSymbol byte, containSymbol bool) ([]byte, error) {
	start := strings.IndexByte(data, startSymbol)
	if start == -1 {
		return nil, ErrCannotParseData
	}

	length := strings.IndexByte(data[start+1:], endSymbol)
	if length == -1 {
		return nil, ErrCannotParseData
	}

	if containSymbol {
		length += 2
	} else {
		start++
	}

	res := make([]byte, length)
	copy(res, data[start:]) // copy the sub string from data to res

	return res, nil
}
