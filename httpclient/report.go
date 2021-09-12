package httpclient

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
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

const reportDomain = "form.hhu.edu.cn"

var (
	prefixArray = [...]string{"var _selfFormWid", "fillDetail"}
	symbolArray = [...]htmlSymbol{symbolString, symbolJSON}
	//ErrCannotParseData cannot parse html data error
	ErrCannotParseData = errors.New("data: parse error")
)

// getFormSessionID 获取打卡系统的SessionID
func (c *punchClient) getFormSessionID() error {
	req, err := getWithContext(c.ctx, "http://"+reportDomain+"/pdc/form/list")
	if err != nil {
		return err
	}

	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return err
	}
	drainBody(res.Body)

	if c.jar.GetCookieByDomain(reportDomain) == nil {
		err = CookieNotFoundErr{"JSESSIONID"}
	}
	return err
}

// getFormDetail 获取打卡表单详细信息
func (c *punchClient) getFormDetail() (form *HealthForm, params *QueryParam, err error) {
	var req *http.Request
	req, err = getWithContext(c.ctx, "http://"+reportDomain+"/pdc/formDesignApi/S/gUTwwojq")
	if err != nil {
		return
	}

	var res *http.Response
	if res, err = c.httpClient.Do(req); err != nil {
		return
	}

	var (
		bufferReader = bufio.NewReader(res.Body)
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
	drainBody(res.Body)

	if err != nil || index != 2 {
		err = ErrCannotParseData
		return
	}

	tmpForm := &HealthForm{}

	if err = json.Unmarshal(resData[indexHealthFormData], tmpForm); err != nil {
		return
	}

	tmpForm.DataTime = time.Now().In(timeZone).Format("2006/01/02") // 表单中增加打卡日期

	if err = fieldZeroValueCheck(tmpForm); err != nil {
		return
	}

	form = tmpForm
	params = &QueryParam{
		Wid:    string(resData[indexWID]),
		UserID: form.StudentID,
	}
	return
}

// postForm 提交打卡表单
func (c *punchClient) postForm(form *HealthForm, params *QueryParam) error {
	value, err := query.Values(form)
	if err != nil {
		return err
	}

	var req *http.Request
	req, err = postFormWithContext(c.ctx,
		"http://"+reportDomain+"/pdc/formDesignApi/dataFormSave",
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

func fieldZeroValueCheck(item interface{}) error {
	v := reflect.ValueOf(item).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			return errors.New("field: '" + v.Type().Field(i).Name + "' is zero value")
		}
	}
	return nil
}
