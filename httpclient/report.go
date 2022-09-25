package httpclient

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrCouldNotGetFormSession get form session id failed
var ErrCouldNotGetFormSession = errors.New("could not get form session")

type htmlSymbol uint8

const (
	symbolJSON htmlSymbol = iota
	symbolString
)

type formValue string

var _ json.Unmarshaler = (*formValue)(nil)

func (v *formValue) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		data = data[1 : len(data)-1]
	}
	*v = formValue(data)
	return nil
}

const reportDomain = "dailyreport.hhu.edu.cn"

var (
	//ErrCannotParseData cannot parse html data error
	ErrCannotParseData = errors.New("data: parse error")

	timeZone = time.FixedZone("CST", 8*3600)
)

// getFormSessionID 获取打卡系统的SessionID
func (c *punchClient) getFormSessionID(retry uint8) (err error) {
	var req *http.Request
	req, err = getWithContext(c.ctx, "http://"+reportDomain+"/pdc/form/list")
	if err != nil {
		return
	}

	var resp *http.Response
	if resp, err = c.retryGet(req, retry); err != nil {
		return
	}
	drainBody(resp.Body)

	if c.httpClient.Jar.Cookies(&url.URL{Host: reportDomain}) == nil {
		err = ErrCouldNotGetFormSession
	}
	return
}

// getFormDetail 获取打卡表单详细信息
func (c *punchClient) getFormDetail(retry uint8) (form map[string]formValue, query string, err error) {
	var req *http.Request
	req, err = getWithContext(c.ctx, "http://"+reportDomain+"/pdc/formDesignApi/S/xznuPIjG")
	if err != nil {
		return
	}

	var resp *http.Response
	if resp, err = c.retryGet(req, retry); err != nil {
		return
	}

	var (
		bufferReader  = bufio.NewReader(resp.Body)
		wid, formData []byte
		line          string
	)

	for err == nil {
		line, err = scanLine(bufferReader)
		if strings.HasPrefix(line, "var _selfFormWid") {
			wid, err = parseData(line, symbolString)
			break
		}
	}
	for err == nil {
		line, err = scanLine(bufferReader)
		if strings.HasPrefix(line, "fillDetail") {
			formData, err = parseData(line, symbolJSON)
			break
		}
	}
	drainBody(resp.Body)

	if err != nil {
		err = fmt.Errorf("get form data failed, err: %w", err)
		return
	}

	tmpForm := make(map[string]formValue)
	if err = json.Unmarshal(formData, &tmpForm); err != nil {
		return
	}

	if err = zeroValueCheck(tmpForm); err != nil {
		return
	}

	query = fmt.Sprintf("wid=%s&userId=%s", string(wid), tmpForm["USERID"])

	delete(tmpForm, "CLRQ")   // 删除填报时间字段
	delete(tmpForm, "USERID") // 删除UserID字段
	delete(tmpForm, "RN")

	tmpForm["DATETIME_CYCLE"] = formValue(time.Now().In(timeZone).Format("2006/01/02")) // 表单中增加打卡日期
	form = tmpForm

	return
}

// postForm 提交打卡表单
func (c *punchClient) postForm(form map[string]formValue, query string) error {
	value := make(url.Values, len(form))
	for key, val := range form {
		value.Set(key, string(val))
	}

	req, err := postFormWithContext(c.ctx,
		"http://"+reportDomain+"/pdc/formDesignApi/dataFormSave",
		value,
	)
	if err != nil {
		return err
	}

	req.URL.RawQuery = query

	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return err
	}
	drainBody(res.Body)

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
		if err != nil {
			res, err = getSlice(data, '"', '"', false)
		}
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

func zeroValueCheck(item map[string]formValue) error {
	if len(item) == 0 {
		return errors.New("check: the map is empty")
	}
	for key, value := range item {
		if value == "" {
			return errors.New("check: '" + key + "' has zero value")
		}
	}
	return nil
}
